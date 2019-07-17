package controller

import (
	"encoding/json"
	"fmt"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/guregu/null"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/ephemeralkey"
	"github.com/stripe/stripe-go/plan"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
	"net/http"
)

var stripeTestPlans = map[string]string{
	"standard_year":  "plan_FOdfeaqzczp6Ag",
	"standard_month": "plan_FOdgPTznDwHU4i",
	"premium_year":   "plan_FOde0uAr0V4WmT",
}

var stripeLivePlans = map[string]string{}

// In sandbox and dev env, use stripeTestPlans.
func getStripePlanID(key string) (string, error) {
	// Only in non-sandbox production env should we use
	// the live key.
	live := Production && !Sandbox

	var id string
	var ok bool

	if live {
		id, ok = stripeLivePlans[key]
	} else {
		id, ok = stripeTestPlans[key]
	}

	if !ok {
		return id, fmt.Errorf("plan for %s not found", key)
	}

	return id, nil
}

type StripeRouter struct {
	PayRouter
}

func NewStripeRouter(m model.Env, p postoffice.Postman, sandbox bool) StripeRouter {
	r := StripeRouter{}

	r.sandbox = sandbox
	r.model = m
	r.postman = p

	return r
}

func stripeBadRequest(err error) view.Response {
	if stripeErr, ok := err.(*stripe.Error); ok {
		resp := view.NewResponse()
		resp.StatusCode = stripeErr.HTTPStatusCode
		resp.Body = stripeErr
		return resp
	}

	return view.NewBadRequest(err.Error())
}

func stripeDBFailure(err error) view.Response {
	if stripeErr, ok := err.(*stripe.Error); ok {
		resp := view.NewResponse()
		resp.StatusCode = stripeErr.HTTPStatusCode
		resp.Body = stripeErr
		return resp
	}

	return view.NewDBFailure(err)
}

// GetPlan retrieves a stripe plan.
func (router StripeRouter) GetPlan(w http.ResponseWriter, req *http.Request) {
	id, err := GetURLParam(req, "id").ToString()
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	planID, err := getStripePlanID(id)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	p, err := plan.Get(planID, nil)
	if err != nil {
		view.Render(w, stripeBadRequest(err))
		return
	}

	view.Render(w, view.NewResponse().SetBody(p))
}

// CreateCustomer send client stripe's customer id.
func (router StripeRouter) CreateCustomer(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(ftcIDKey)

	stripeID, err := router.model.CreateStripeCustomer(ftcID)

	if err != nil {
		view.Render(w, stripeDBFailure(err))
		return
	}

	view.Render(w, view.NewResponse().SetBody(map[string]string{
		"id": stripeID,
	}))
}

// GetDefaultPaymentMethod retrieves a user's invoice default payment method.
func (router StripeRouter) GetDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	id, err := GetURLParam(req, "id").ToString()
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	c, err := customer.Get(id, nil)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	pm := c.InvoiceSettings.DefaultPaymentMethod
	if pm == nil {
		view.Render(w, view.NewNotFound())
		return
	}

	view.Render(w, view.NewResponse().SetBody(pm))
}

// SetDefaultPaymentMethod sets stripe customer's invoice_settings.default_payment_method.
func (router StripeRouter) SetDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	id, err := GetURLParam(req, "id").ToString()
	if err != nil {
		logrus.Error(err)
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	pmID, err := util.GetJSONString(req.Body, "defaultPaymentMethod")
	if err != nil {
		logrus.Error(err)
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pmID),
		},
	}

	c, err := customer.Update(id, params)

	if err != nil {
		logrus.Error(err)
		view.Render(w, stripeBadRequest(err))
		return
	}

	view.Render(w, view.NewResponse().SetBody(c))
}

// IssueKey creates an ephemeral key.
//
// GET /stripe/customers/:id/key?api_version=<version>
func (router StripeRouter) IssueKey(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	id, err := GetURLParam(req, "id").ToString()
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	stripeVersion := req.Form.Get("api_version")
	if stripeVersion == "" {
		view.Render(w, view.NewBadRequest("Stripe-Version not found"))
		return
	}

	params := &stripe.EphemeralKeyParams{
		Customer:      stripe.String(id),
		StripeVersion: stripe.String(stripeVersion),
	}
	key, err := ephemeralkey.New(params)
	if err != nil {
		view.Render(w, stripeBadRequest(err))
		return
	}

	w.Write(key.RawJSON)
}

// CreateSubscription create a stripe subscription.
// Input: {customer: "", coupon: "", defaultPaymentMethod: "", ftcPlanId: "standard_year" | "standard_month" | "premium_year"}
//
// PITFALLS:
// If you creates a plan in CNY, and a customer is subscribed to
// it, and after that you created another plan in GBP, then
// Stripe will decline your subsequent subscription request.
// It's better to create different plans in the same currency.
// I guess Stripe takes plans in different currencies as the
// same one to avoid customer subscribing to the same plan
// in different countries and regions.
// {
// "status":400,
// "message":"You cannot combine currencies on a single customer. This customer has had a subscription, coupon, or invoice item with currency cny",
// "request_id":"req_fa0rfmytgnI22E",
// "type":"invalid_request_error"
// }
func (router StripeRouter) CreateSubscription(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(ftcIDKey)
	// "plan_FOEFa7c1zLOtJW"
	var params paywall.StripeSubParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	logrus.WithField("trace", "StripeRouter").Infof("Stripe param: %+v", params)

	planID, err := getStripePlanID(params.Key())
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	params.PlanID = planID
	s, err := router.model.CreateStripeSub(paywall.UserID{
		CompoundID: ftcID,
		FtcID:      null.StringFrom(ftcID),
		UnionID:    null.String{},
	}, params)

	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	view.Render(w, view.NewResponse().SetBody(s))
}

// GetSubscription fetches a user's subscription and update membership if data in our db is stale.
func (router StripeRouter) GetSubscription(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(ftcIDKey)

	id, _ := paywall.NewUserID(null.StringFrom(ftcID), null.String{})

	ss, err := router.model.GetStripeSub(id)
	if err != nil {
		view.Render(w, stripeDBFailure(err))
		return
	}

	view.Render(w, view.NewResponse().SetBody(ss))
}

// UpgradeSubscription create a stripe subscription.
// Input: {customer: "", coupon: "", defaultPaymentMethod: "", ftcPlanId: "standard_year" | "standard_month" | "premium_year"}
func (router StripeRouter) UpgradeSubscription(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(ftcIDKey)
	// "plan_FOEFa7c1zLOtJW"
	var params paywall.StripeSubParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	if params.Key() == "standard_month" {
		view.Render(w, view.NewBadRequest("This is used only to change monthly pan to yearly or from standard to premium."))
		return
	}

	planID, err := getStripePlanID(params.Key())
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	params.PlanID = planID
	s, err := router.model.UpgradeStripeSubs(paywall.UserID{
		CompoundID: ftcID,
		FtcID:      null.StringFrom(ftcID),
		UnionID:    null.String{},
	}, params)

	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	view.Render(w, view.NewResponse().SetBody(s))
}

func (router StripeRouter) WebHook(w http.ResponseWriter, req *http.Request) {
	event := stripe.Event{}

	if err := gorest.ParseJSON(req.Body, &event); err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	logrus.WithField("trace", "StripeRouter.WebHook").Infof("Stripe event type: %s", event.Type)

	invoice := stripe.Invoice{}
	switch event.Type {
	case "customer.subscription.created":
		s := stripe.Subscription{}
		if err := json.Unmarshal(event.Data.Raw, &s); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logrus.Info(s)
		w.WriteHeader(http.StatusOK)
		return

	case "customer.subscription.updated":
		s := stripe.Subscription{}
		if err := json.Unmarshal(event.Data.Raw, &s); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logrus.Info(s)
		w.WriteHeader(http.StatusOK)
		return

	case "invoice.created":
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	case "invoice.finalized":
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logrus.Info(invoice)
		w.WriteHeader(http.StatusOK)
		return

	case "invoice.payment_failed":
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logrus.Info(invoice)
		w.WriteHeader(http.StatusOK)
		return

	case "invoice.payment_succeeded":
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logrus.Info(invoice)
		w.WriteHeader(http.StatusOK)
		return

	case "invoice.upcoming":
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logrus.Info(invoice)
		w.WriteHeader(http.StatusOK)
		return
	}

	view.Render(w, view.NewNoContent())
}

// CreatePayIntent for stripe and returns the client secret.
// Client should already created an subscription order before hitting this endpoint.
// Client send the order id here and server uses the id to retrieve the order details, like price, to ask Stripe to create
// a payment intent.
// See: https://stripe.com/docs/payments/payment-intents/android
//
// Input: {orderId: "FTxxxxxx"}
//
// Output: {secret: "xxxxx"}
//func (router StripeRouter) CreatePayIntent(w http.ResponseWriter, req *http.Request) {
//	ftcID := req.Header.Get(ftcIDKey)
//	orderID, err := util.GetJSONString(req.Body, "orderId")
//	if err != nil {
//		view.Render(w, view.NewBadRequest(err.Error()))
//		return
//	}
//	if orderID == "" {
//		reason := view.NewReason()
//		reason.Field = "orderId"
//		reason.Code = view.CodeMissingField
//		reason.SetMessage("Order id must be provided")
//		view.Render(w, view.NewUnprocessable(reason))
//		return
//	}
//
//	ftcUser, err := router.model.FindFtcUser(ftcID)
//	if err != nil {
//		view.Render(w, view.NewDBFailure(err))
//		return
//	}
//	if ftcUser.StripeID.IsZero() {
//		view.Render(w, view.NewNotFound())
//		return
//	}
//
//	billing, err := router.model.FindOrderBilling(orderID, ftcID)
//	if err != nil {
//		view.Render(w, view.NewDBFailure(err))
//		return
//	}
//	if billing.IsConfirmed {
//		view.Render(w, view.NewForbidden("The order provided is already confirmed."))
//		return
//	}
//
//	intent, err := stripepay.CreatePaymentIntent(billing.PriceInCent(), ftcUser.StripeID.String)
//	if err != nil {
//		view.Render(w, view.NewBadRequest(err.Error()))
//		return
//	}
//
//	view.Render(w, view.NewResponse().SetBody(map[string]string{
//		"secret": intent.ClientSecret,
//	}))
//}
