package controller

import (
	"encoding/json"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/view"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	stripesdk "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/ephemeralkey"
	stripeplan "github.com/stripe/stripe-go/plan"
	"github.com/stripe/stripe-go/webhook"
	ftcplan "gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	ftcstripe "gitlab.com/ftchinese/subscription-api/models/stripe"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/pkg/config"
	"gitlab.com/ftchinese/subscription-api/repository/readerrepo"
	"gitlab.com/ftchinese/subscription-api/repository/striperepo"
	"io/ioutil"
	"net/http"
)

type StripeRouter struct {
	signingKey string
	readerEnv  readerrepo.ReaderEnv
	stripeEnv  striperepo.StripeEnv
}

// NewStripeRouter initializes StripeRouter.
func NewStripeRouter(db *sqlx.DB, config config.BuildConfig) StripeRouter {
	r := StripeRouter{
		signingKey: config.GetStripeKey(),
		readerEnv:  readerrepo.NewReaderEnv(db, config),
		stripeEnv:  striperepo.NewStripeEnv(db, config),
	}

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

// GetPlan retrieves a stripe plan by id.
// GET /stripe/plans/<standard_month | standard_year | premium_year>
func (router StripeRouter) GetPlan(w http.ResponseWriter, req *http.Request) {
	key, err := GetURLParam(req, "id").ToString()
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	ftcPlan, err := ftcplan.GetPlans().FindPlan(key)

	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	p, err := stripeplan.Get(
		ftcPlan.GetStripePlanID(router.stripeEnv.Live()),
		nil)

	if err != nil {
		_ = view.Render(w, stripeBadRequest(err))
		return
	}

	_ = view.Render(w, view.NewResponse().SetBody(p))
}

// CreateCustomer send client stripe's customer id.
// PUT /stripe/customers
func (router StripeRouter) CreateCustomer(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(ftcIDKey)

	account, err := router.stripeEnv.CreateStripeCustomer(ftcID)

	if err != nil {
		_ = view.Render(w, stripeDBFailure(err))
		return
	}

	data := struct {
		ID string `json:"id"` // Deprecated. Reserved for backward compatibility. Should be removed after Android version 3.3.0
		reader.Account
	}{
		ID:      account.StripeID.String,
		Account: account,
	}
	_ = view.Render(w, view.NewResponse().SetBody(data))
}

// GetDefaultPaymentMethod retrieves a user's invoice default payment method.
// GET /stripe/customers//{id}/default_payment_method
func (router StripeRouter) GetDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	// Get stripe customer id from url.
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
		_ = view.Render(w, view.NewNotFound())
		return
	}

	_ = view.Render(w, view.NewResponse().SetBody(pm))
}

// SetDefaultPaymentMethod sets stripe customer's invoice_settings.default_payment_method.
// POST /stripe/customers//{id}/default_payment_method
//
// Input: {defaultPaymentMethod: string}. The id the the default payment method.
func (router StripeRouter) SetDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	// Get stripe customer id from url.
	id, err := GetURLParam(req, "id").ToString()
	if err != nil {
		logrus.Error(err)
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	pmID, err := util.GetJSONString(req.Body, "defaultPaymentMethod")
	if err != nil {
		logrus.Error(err)
		_ = view.Render(w, view.NewBadRequest(err.Error()))
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
		_ = view.Render(w, stripeBadRequest(err))
		return
	}

	_ = view.Render(w, view.NewResponse().SetBody(c))
}

// IssueKey creates an ephemeral key.
//
// GET /stripe/customers/{id}/key?api_version=<version>
func (router StripeRouter) IssueKey(w http.ResponseWriter, req *http.Request) {
	log := logrus.WithField("trace", "StripeRouter.IssueKey")

	if err := req.ParseForm(); err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// Get stripe customer id.
	id, err := GetURLParam(req, "id").ToString()
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	stripeVersion := req.Form.Get("api_version")
	if stripeVersion == "" {
		_ = view.Render(w, view.NewBadRequest("Stripe-Version not found"))
		return
	}

	params := &stripesdk.EphemeralKeyParams{
		Customer:      stripesdk.String(id),
		StripeVersion: stripesdk.String(stripeVersion),
	}
	key, err := ephemeralkey.New(params)
	if err != nil {
		_ = view.Render(w, stripeBadRequest(err))
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
// in case user already linked wechat.
// Notification email is sent upon webhook receiving data, not here.
func (router StripeRouter) CreateSubscription(w http.ResponseWriter, req *http.Request) {
	log := logrus.WithField("trace", "StripeRouter.CreateSubscription")

	// Get FTC id.
	ftcID, _ := GetUserID(req.Header)

	// "plan_FOEFa7c1zLOtJW"
	var params ftcstripe.SubParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}
	params.SetLive(router.stripeEnv.Live())

	log.Infof("Stripe param: %+v", params)

	// Create stripe subscription.
	s, err := router.stripeEnv.CreateSubscription(ftcID, params)

	if err != nil {
		if sErr := CastStripeError(err); sErr != nil {
			_ = view.Render(w, BuildStripeResponse(sErr))

			go func() {
				_ = router.stripeEnv.SaveSubsError(ftcID, sErr)
			}()
			return
		}

		switch err {
		case util.ErrNonStripeValidSub,
			util.ErrActiveStripeSub,
			util.ErrUnknownSubState:
			_ = view.Render(w, view.NewBadRequest(err.Error()))
		default:
			_ = view.Render(w, view.NewDBFailure(err))
		}
		return
	}

	resp, err := ftcstripe.NewPaymentResult(s)
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	log.Infof("Subscription id %s, status %s, payment intent status %s", s.ID, s.Status, s.LatestInvoice.PaymentIntent.Status)

	_ = view.Render(w, view.NewResponse().SetBody(resp))
}

// GetSubscription fetches a user's subscription and update membership if data in our db is stale.
//
// Error Response:
// 404: membership for this user is not found.
func (router StripeRouter) GetSubscription(w http.ResponseWriter, req *http.Request) {
	log := logger.WithField("trace", "StripeRouter.GetSubscription")

	userID, _ := GetUserID(req.Header)

	s, err := router.stripeEnv.GetSubscription(userID)
	if err != nil {
		log.Error(err)

		if sErr := CastStripeError(err); sErr != nil {
			_ = view.Render(w, BuildStripeResponse(sErr))
			return
		}

		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	log.Infof("Subscription id %s, status %s", s.ID, s.Status)

	_ = view.Render(w, view.NewResponse().SetBody(ftcstripe.NewSubsResponse(s)))
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
	var params ftcstripe.SubParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}
	params.SetLive(router.stripeEnv.Live())

	log.Infof("Subscription params: %+v", params)

	s, err := router.stripeEnv.UpgradeSubscription(userID, params)

	if err != nil {
		if sErr := CastStripeError(err); sErr != nil {
			_ = view.Render(w, BuildStripeResponse(sErr))

			go func() {
				err := router.stripeEnv.SaveSubsError(userID, sErr)
				if err != nil {
					log.Error(err)
				}
			}()
			return
		}

		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	resp, err := ftcstripe.NewPaymentResult(s)
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	log.Infof("Subscription id %s, status %s, payment intent status %s", s.ID, s.Status, s.LatestInvoice.PaymentIntent.Status)

	_ = view.Render(w, view.NewResponse().SetBody(resp))
}

func (router StripeRouter) onSubscription(s *stripesdk.Subscription) error {

	account, err := router.readerEnv.FindAccountByStripeID(s.Customer.ID)
	if err != nil {
		return err
	}

	memberID := account.MemberID()

	return router.stripeEnv.WebHookOnSubscription(memberID, s)
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
//		parcel, err = ftcUser.StripeInvoiceParcel(paywall.Invoice{i})
//	case invoiceEventFailed:
//		parcel, err = ftcUser.StripePaymentFailed(paywall.Invoice{i})
//	case invoiceEventRequiresAction:
//		parcel, err = ftcUser.StripeActionRequired(paywall.Invoice{i})
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
