package controller

import (
	"encoding/json"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/sirupsen/logrus"
	stripesdk "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/ephemeralkey"
	stripeplan "github.com/stripe/stripe-go/plan"
	"github.com/stripe/stripe-go/webhook"
	ftcplan "gitlab.com/ftchinese/subscription-api/models/plan"
	ftcstripe "gitlab.com/ftchinese/subscription-api/models/stripe"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository/subrepo"
	"io/ioutil"
	"net/http"
)

type StripeRouter struct {
	signingKey string
	PayRouter
}

func NewStripeRouter(m subrepo.SubEnv, p postoffice.Postman, sk string) StripeRouter {
	r := StripeRouter{
		signingKey: sk,
	}

	r.subEnv = m
	r.postman = p

	return r
}

func stripeBadRequest(err error) view.Response {
	if stripeErr, ok := err.(*stripesdk.Error); ok {
		resp := view.NewResponse()
		resp.StatusCode = stripeErr.HTTPStatusCode
		resp.Body = stripeErr
		return resp
	}

	return view.NewBadRequest(err.Error())
}

func stripeDBFailure(err error) view.Response {
	if stripeErr, ok := err.(*stripesdk.Error); ok {
		resp := view.NewResponse()
		resp.StatusCode = stripeErr.HTTPStatusCode
		resp.Body = stripeErr
		return resp
	}

	return view.NewDBFailure(err)
}

// GetPlan retrieves a stripe plan.
// GET /stripe/plans/<standard_month | standard_year | premium_year>
func (router StripeRouter) GetPlan(w http.ResponseWriter, req *http.Request) {
	key, err := GetURLParam(req, "id").ToString()
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	ftcPlan, err := ftcplan.FindFtcPlan(key)

	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	p, err := stripeplan.Get(
		ftcPlan.GetStripePlanID(router.subEnv.Live()),
		nil)

	if err != nil {
		_ = view.Render(w, stripeBadRequest(err))
		return
	}

	_ = view.Render(w, view.NewResponse().SetBody(p))
}

// CreateCustomer send client stripe's customer id.
func (router StripeRouter) CreateCustomer(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(ftcIDKey)

	account, err := router.subEnv.CreateStripeCustomer(ftcID)

	if err != nil {
		_ = view.Render(w, stripeDBFailure(err))
		return
	}

	_ = view.Render(w, view.NewResponse().SetBody(map[string]string{
		"id": account.StripeID.String,
	}))
}

// GetDefaultPaymentMethod retrieves a user's invoice default payment method.
func (router StripeRouter) GetDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	id, err := GetURLParam(req, "id").ToString()
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	c, err := customer.Get(id, nil)
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
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

	params := &stripesdk.CustomerParams{
		InvoiceSettings: &stripesdk.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripesdk.String(pmID),
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
	log := logrus.WithField("trace", "StripeRouter.IssueKey")

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

	params := &stripesdk.EphemeralKeyParams{
		Customer:      stripesdk.String(id),
		StripeVersion: stripesdk.String(stripeVersion),
	}
	key, err := ephemeralkey.New(params)
	if err != nil {
		view.Render(w, stripeBadRequest(err))
		return
	}

	_, err = w.Write(key.RawJSON)
	if err != nil {
		log.Error(err)
	}
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
// in case user alread linked wechat.
// Notification email is sent upon webhook receiving data, not here.
func (router StripeRouter) CreateSubscription(w http.ResponseWriter, req *http.Request) {
	log := logrus.WithField("trace", "StripeRouter")

	userID, _ := GetUserID(req.Header)

	// "plan_FOEFa7c1zLOtJW"
	var params ftcstripe.StripeSubParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// Attach Stripe plan id.
	err := params.SetStripePlanID(router.subEnv.Live())
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	log.Infof("Stripe param: %+v", params)

	// Create stripe subscription.
	s, err := router.subEnv.CreateStripeSub(userID, params)

	if err != nil {
		if sErr := CastStripeError(err); sErr != nil {
			view.Render(w, BuildStripeResponse(sErr))

			go func() {
				err := router.subEnv.SaveStripeError(userID, sErr)
				if err != nil {
					log.Error(err)
				}
			}()
			return
		}

		switch err {
		case util.ErrNonStripeValidSub,
			util.ErrActiveStripeSub,
			util.ErrUnknownSubState:
			view.Render(w, view.NewBadRequest(err.Error()))
		default:
			view.Render(w, view.NewDBFailure(err))
		}
		return
	}

	resp, err := ftcstripe.BuildStripeSubResponse(s)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	log.Infof("Subscription id %s, status %s, payment intent status %s", s.ID, s.Status, s.LatestInvoice.PaymentIntent.Status)

	view.Render(w, view.NewResponse().SetBody(resp))
}

// GetSubscription fetches a user's subscription and update membership if data in our db is stale.
//
// Error Response:
// 404: membership for this user is not found.
func (router StripeRouter) GetSubscription(w http.ResponseWriter, req *http.Request) {
	log := logrus.WithField("trace", "StripeRouter.GetSubscription")

	userID, _ := GetUserID(req.Header)

	s, err := router.subEnv.GetStripeSub(userID)
	if err != nil {
		logrus.WithField("trace", "StripeRouter.GetSubscription").Error(err)

		if sErr := CastStripeError(err); sErr != nil {
			view.Render(w, BuildStripeResponse(sErr))
			return
		}

		view.Render(w, view.NewDBFailure(err))
		return
	}

	log.Infof("Subscription id %s, status %s", s.ID, s.Status)

	view.Render(w, view.NewResponse().SetBody(ftcstripe.NewStripeSub(s)))
}

// UpgradeSubscription create a stripe subscription.
// Input: {customer: "", coupon: "", defaultPaymentMethod: "", ftcPlanId: "standard_year" | "standard_month" | "premium_year"}
//
// Error response:
// 404 if membership if not found.
// NOTE: when updating a stripe subscription, the return payload
// `items` field contains more than one items:
// one is standard and another if premium.
// So we cannot rely on this field to find FTC plan.
func (router StripeRouter) UpgradeSubscription(w http.ResponseWriter, req *http.Request) {
	log := logrus.WithField("trace", "StripeRouter.UpgradeSubscription")

	userID, _ := GetUserID(req.Header)

	// "plan_FOEFa7c1zLOtJW"
	var params ftcstripe.StripeSubParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	err := params.SetStripePlanID(router.subEnv.Live())
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	log.Infof("Subscription params: %+v", params)

	s, err := router.subEnv.UpgradeStripeSubs(userID, params)

	if err != nil {
		if sErr := CastStripeError(err); sErr != nil {
			view.Render(w, BuildStripeResponse(sErr))

			go func() {
				err := router.subEnv.SaveStripeError(userID, sErr)
				if err != nil {
					log.Error(err)
				}
			}()
			return
		}

		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	resp, err := ftcstripe.BuildStripeSubResponse(s)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	log.Infof("Subscription id %s, status %s, payment intent status %s", s.ID, s.Status, s.LatestInvoice.PaymentIntent.Status)

	_ = view.Render(w, view.NewResponse().SetBody(resp))
}

func (router StripeRouter) onSubscription(s *stripesdk.Subscription) error {
	logger := logrus.WithField("trace", "StripeRouter.onSubscription")

	_, err := router.subEnv.WebHookSaveStripeSub(s)

	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

//func (router StripeRouter) onInvoice(i *stripe.Invoice, event invoiceEvent) error {
//	logger := logrus.WithField("trace", "StripeRouter.onInvoice")
//
//	ftcUser, err := router.env.FindStripeCustomer(i.Customer.ID)
//	if err != nil {
//		return err
//	}
//
//	var parcel postoffice.Parcel
//	switch event {
//	case invoiceEventSucceeded:
//		parcel, err = ftcUser.StripeInvoiceParcel(paywall.StripeInvoice{i})
//	case invoiceEventFailed:
//		parcel, err = ftcUser.StripePaymentFailed(paywall.StripeInvoice{i})
//	case invoiceEventRequiresAction:
//		parcel, err = ftcUser.StripeActionRequired(paywall.StripeInvoice{i})
//	}
//
//	if err != nil {
//		logger.Error(err)
//		return err
//	}
//
//	err = router.postman.Deliver(parcel)
//	if err != nil {
//		logger.Error(err)
//		return err
//	}
//
//	return nil
//}

func (router StripeRouter) WebHook(w http.ResponseWriter, req *http.Request) {
	logger := logrus.WithField("trace", "StripeRouter.WebHook")

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Errorf("Error reading request body %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := webhook.ConstructEvent(body, req.Header.Get("Stripe-Signature"), router.signingKey)
	if err != nil {
		logger.Errorf("Error verifying webhook signature: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Infof("Stripe event received: %s", event.Type)

	switch event.Type {

	// Occurs whenever a customer is signed up for a new plan.
	case "customer.subscription.created":
		s := stripesdk.Subscription{}
		if err := json.Unmarshal(event.Data.Raw, &s); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Info(s)

		w.WriteHeader(http.StatusOK)

		go func() {
			err := router.onSubscription(&s)
			if err != nil {
				logger.Error(err)
			}
		}()
		return

	//	Occurs whenever a subscription changes (e.g., switching from one plan to another, or changing the status from trial to active).
	case "customer.subscription.updated":
		s := stripesdk.Subscription{}
		if err := json.Unmarshal(event.Data.Raw, &s); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Info(s)
		w.WriteHeader(http.StatusOK)

		go func() {
			err := router.onSubscription(&s)
			if err != nil {
				logger.Error(err)
			}
		}()

		return

	case "invoice.created":
		// Stripe waits an hour after receiving a successful response to the invoice.created event before attempting payment.
		// If a successful response isn’t received within 72 hours, Stripe attempts to finalize and send the invoice.
		// In live mode, if your webhook endpoint does not respond properly, Stripe continues retrying the webhook notification for up to three days with an exponential back off
		var i stripesdk.Invoice
		if err := json.Unmarshal(event.Data.Raw, &i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return

	// Handling payment failures
	case "invoice.payment_failed":
		// Send email to user.
		var i stripesdk.Invoice
		if err := json.Unmarshal(event.Data.Raw, &i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		//go func() {
		//	err := router.onInvoice(&i, invoiceEventFailed)
		//	if err != nil {
		//		logger.Error(err)
		//	}
		//}()
		return

	// Handling payments that require additional action
	case "invoice.payment_action_required":
		// Send email to user.
		// Send email to user.
		var i stripesdk.Invoice
		if err := json.Unmarshal(event.Data.Raw, &i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		//go func() {
		//	err := router.onInvoice(&i, invoiceEventRequiresAction)
		//	if err != nil {
		//		logger.Error(err)
		//	}
		//}()
		return

	// Tracking active subscriptions
	// A few days prior to renewal, your site receives an invoice.upcoming event at the webhook endpoint.
	case "invoice.upcoming":
		var i stripesdk.Invoice
		if err := json.Unmarshal(event.Data.Raw, &i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return

	case "invoice.payment_succeeded":
		// Send email to user.
		var i stripesdk.Invoice
		if err := json.Unmarshal(event.Data.Raw, &i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		//go func() {
		//	err := router.onInvoice(&i, invoiceEventSucceeded)
		//	if err != nil {
		//		logger.Error(err)
		//	}
		//}()
		return

	case "invoice.finalized":
		var i stripesdk.Invoice
		if err := json.Unmarshal(event.Data.Raw, &i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}

//type invoiceEvent int
//
//const (
//	invoiceEventSucceeded invoiceEvent = iota
//	invoiceEventFailed
//	invoiceEventRequiresAction
//)
