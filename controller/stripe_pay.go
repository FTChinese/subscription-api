package controller

import (
	"encoding/json"
	"errors"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/reader"
	stripePkg "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/FTChinese/subscription-api/repository/readerrepo"
	"github.com/FTChinese/subscription-api/repository/striperepo"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	stripeSdk "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
	"io/ioutil"
	"net/http"
)

type StripeRouter struct {
	config     config.BuildConfig
	signingKey string
	readerRepo readerrepo.Env
	stripeRepo striperepo.StripeEnv
}

// NewStripeRouter initializes StripeRouter.
func NewStripeRouter(db *sqlx.DB, config config.BuildConfig) StripeRouter {
	r := StripeRouter{
		config:     config,
		signingKey: config.MustStripeSigningKey(),
		readerRepo: readerrepo.NewEnv(db, config),
		stripeRepo: striperepo.NewStripeEnv(db, config),
	}

	return r
}

// Forward stripe error to client, and give the error back to caller to handle if it is not stripe error.
func forwardStripeErr(w http.ResponseWriter, err error) error {

	if stripeErr, ok := err.(*stripeSdk.Error); ok {
		return render.New(w).
			JSON(stripeErr.HTTPStatusCode, stripeErr)
	}

	var ve *render.ValidationError
	if errors.As(err, &ve) {
		return render.New(w).Unprocessable(ve)
	}

	return err
}

// GetPlan retrieves a stripe plan by id.
// GET /stripe/plans/<standard_month | standard_year | premium_year>
func (router StripeRouter) GetPlan(w http.ResponseWriter, req *http.Request) {
	key, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Fetch plan from Stripe API
	p, err := stripePkg.FetchPlan(key, router.config.Live())

	if err != nil {
		err = forwardStripeErr(w, err)
		if err == nil {
			return
		}

		_ = render.New(w).NotFound()
		return
	}

	_ = render.New(w).OK(p)
}

// CreateCustomer creates stripe customer if not present.
// PUT /stripe/customers
// Response: reader.Account
func (router StripeRouter) CreateCustomer(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(ftcIDKey)

	account, err := router.stripeRepo.CreateStripeCustomer(ftcID)

	if err != nil {
		err := forwardStripeErr(w, err)
		if err != nil {
			_ = render.New(w).DBError(err)
		}

		return
	}

	_ = render.New(w).OK(account)
}

// GetDefaultPaymentMethod retrieves a user's invoice default payment method.
// GET /stripe/customers/{id}/default_payment_method
func (router StripeRouter) GetDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	// Get stripe customer id from url.
	id, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	pm, err := stripePkg.GetDefaultPaymentMethod(id)
	if err != nil {
		err = forwardStripeErr(w, err)
		if err != nil {
			_ = render.New(w).DBError(err)
		}
		return
	}

	_ = render.New(w).OK(pm)
}

// SetDefaultPaymentMethod sets stripe customer's invoice_settings.default_payment_method.
// POST /stripe/customers/{id}/default_payment_method
//
// Input: {defaultPaymentMethod: string}. The id of the default payment method.
func (router StripeRouter) SetDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	// Get stripe customer id from url.
	id, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	var pm stripePkg.PaymentInput
	if err := gorest.ParseJSON(req.Body, &pm); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	pm.CustomerID = id

	if ve := pm.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	cus, err := stripePkg.SetDefaultPaymentMethod(pm)
	if err != nil {
		err = forwardStripeErr(w, err)
		if err != nil {
			_ = render.New(w).BadRequest(err.Error())
		}
		return
	}

	_ = render.New(w).OK(cus)
}

// IssueKey creates an ephemeral key.
//
// GET /stripe/customers/{id}/key?api_version=<version>
func (router StripeRouter) IssueKey(w http.ResponseWriter, req *http.Request) {
	log := logrus.WithField("trace", "StripeRouter.IssueKey")

	// Get stripe customer id.
	id, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	stripeVersion := req.FormValue("api_version")
	if stripeVersion == "" {
		_ = render.New(w).BadRequest("Stripe-Version not found")
		return
	}

	keyData, err := stripePkg.GetEphemeralKey(id, stripeVersion)
	if err != nil {
		err = forwardStripeErr(w, err)
		if err != nil {
			_ = render.New(w).BadRequest(err.Error())
		}
		return
	}

	_, err = w.Write(keyData)
	if err != nil {
		log.Error(err)
	}
}

// CreateSubscription create a stripe subscription.
// Input:
// tier: string;
// cycle: string;
// customer: string;
// coupon?: string;
// defaultPaymentMethod?: string;
// idempotency: string;
// Why this field?
// ftcPlanId: "standard_year" | "standard_month" | "premium_year"
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

	// Get FTC id. Its presence is already checked by middleware.
	ftcID := req.Header.Get(ftcIDKey)

	input := stripePkg.NewSubsInput(ftcID)
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	input, err := input.WithPlanID(router.config.Live())
	if err != nil {
		_ = render.New(w).NotFound()
		return
	}

	// TODO: validate input

	// Create stripe subscription.
	s, err := router.stripeRepo.CreateSubscription(input)

	if err != nil {
		err := forwardStripeErr(w, err)
		if err == nil {
			go func() {
				apiErr := stripePkg.NewAPIError(input.FtcID, castStripeError(err))
				_ = router.stripeRepo.SaveSubsError(apiErr)
			}()

			return
		}

		switch err {
		case reader.ErrNonStripeValidSub,
			reader.ErrActiveStripeSub,
			reader.ErrUnknownSubState:
			_ = render.New(w).BadRequest(err.Error())
		default:
			_ = render.New(w).DBError(err)
		}
		return
	}

	// Tells client whether further action is required.
	resp, err := stripePkg.NewPaymentResult(s)
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	log.Infof("Subscription id %s, status %s, payment intent status %s", s.ID, s.Status, s.LatestInvoice.PaymentIntent.Status)

	_ = render.New(w).OK(resp)
}

// GetSubscription fetches a user's subscription and update membership if data in our db is stale.
//
// Error Response:
// 404: membership for this user is not found.
func (router StripeRouter) GetSubscription(w http.ResponseWriter, req *http.Request) {
	log := logger.WithField("trace", "StripeRouter.GetSubscription")

	readerIDs := getReaderIDs(req.Header)

	s, err := router.stripeRepo.GetSubscription(readerIDs)
	if err != nil {
		log.Error(err)

		err = forwardStripeErr(w, err)
		if err == nil {
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	log.Infof("Subscription id %s, status %s", s.ID, s.Status)

	_ = render.New(w).OK(s)
}

// UpgradeSubscription create a stripe subscription.
// Input:
// tier: string;
// cycle: stirng;
// customer: "",
// coupon?: "",
// defaultPaymentMethod?: "",
//
// ftcPlanId: "standard_year" | "standard_month" | "premium_year"
//
// Error response:
// 404 if membership if not found.
// NOTE: when updating a stripe subscription, the return payload
// `items` field contains more than one items:
// one is standard and another if premium.
// So we cannot rely on this field to find FTC plan.
func (router StripeRouter) UpgradeSubscription(w http.ResponseWriter, req *http.Request) {
	log := logrus.WithField("trace", "StripeRouter.UpgradeSubscription")

	// Get FTC id. Its presence is already checked by middleware.
	ftcID := req.Header.Get(ftcIDKey)

	input := stripePkg.NewSubsInput(ftcID)
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	input, err := input.WithPlanID(router.config.Live())
	if err != nil {
		_ = render.New(w).NotFound()
		return
	}

	// TODO: validate input

	s, err := router.stripeRepo.UpgradeSubscription(input)

	if err != nil {
		err := forwardStripeErr(w, err)
		if err == nil {
			go func() {
				apiErr := stripePkg.NewAPIError(input.FtcID, castStripeError(err))
				_ = router.stripeRepo.SaveSubsError(apiErr)
			}()

			return
		}

		_ = render.New(w).BadRequest(err.Error())
		return
	}

	resp, err := stripePkg.NewPaymentResult(s)
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	log.Infof("Subscription id %s, status %s, payment intent status %s", s.ID, s.Status, s.LatestInvoice.PaymentIntent.Status)

	_ = render.New(w).OK(resp)
}

func (router StripeRouter) onSubscription(s *stripeSdk.Subscription) error {

	account, err := router.readerRepo.AccountByStripeID(s.Customer.ID)
	if err != nil {
		return err
	}

	memberID := account.MemberID()

	return router.stripeRepo.WebHookOnSubscription(memberID, s)
}

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
		s := stripeSdk.Subscription{}
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
		s := stripeSdk.Subscription{}
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
		var i stripeSdk.Invoice
		if err := json.Unmarshal(event.Data.Raw, &i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return

	// Handling payment failures
	case "invoice.payment_failed":
		// Send email to user.
		var i stripeSdk.Invoice
		if err := json.Unmarshal(event.Data.Raw, &i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return

	// Handling payments that require additional action
	case "invoice.payment_action_required":
		// Send email to user.
		// Send email to user.
		var i stripeSdk.Invoice
		if err := json.Unmarshal(event.Data.Raw, &i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return

	// Tracking active subscriptions
	// A few days prior to renewal, your site receives an invoice.upcoming event at the webhook endpoint.
	case "invoice.upcoming":
		var i stripeSdk.Invoice
		if err := json.Unmarshal(event.Data.Raw, &i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return

	case "invoice.payment_succeeded":
		// Send email to user.
		var i stripeSdk.Invoice
		if err := json.Unmarshal(event.Data.Raw, &i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return

	case "invoice.finalized":
		var i stripeSdk.Invoice
		if err := json.Unmarshal(event.Data.Raw, &i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}
