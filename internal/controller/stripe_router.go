package controller

import (
	"encoding/json"
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/readerrepo"
	"github.com/FTChinese/subscription-api/internal/repository/striperepo"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/jmoiron/sqlx"
	stripeSdk "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

type StripeRouter struct {
	config     config.BuildConfig
	signingKey string
	readerRepo readerrepo.Env
	stripeRepo striperepo.Env
	client     striperepo.Client
	logger     *zap.Logger
}

// NewStripeRouter initializes StripeRouter.
func NewStripeRouter(db *sqlx.DB, cfg config.BuildConfig, logger *zap.Logger) StripeRouter {
	client := striperepo.NewClient(cfg.Live(), logger)

	return StripeRouter{
		config:     cfg,
		signingKey: config.MustLoadStripeSigningKey().Pick(cfg.Live()),
		readerRepo: readerrepo.NewEnv(db),
		stripeRepo: striperepo.NewEnv(db, client, logger),
		client:     client,
		logger:     logger,
	}
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

// IssueKey creates an ephemeral key.
//
// GET /stripe/customers/{id}/key?api_version=<version>
func (router StripeRouter) IssueKey(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

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

	keyData, err := router.client.CreateEphemeralKey(id, stripeVersion)
	if err != nil {
		err = forwardStripeErr(w, err)
		if err != nil {
			_ = render.New(w).BadRequest(err.Error())
		}
		return
	}

	_, err = w.Write(keyData)
	if err != nil {
		sugar.Error(err)
	}
}

func (router StripeRouter) onSubscription(s *stripeSdk.Subscription) error {

	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	account, err := router.readerRepo.FtcAccountByStripeID(s.Customer.ID)
	if err != nil {
		sugar.Error(err)
		return err
	}

	memberID := account.MemberID()

	_, err = router.stripeRepo.WebHookOnSubscription(memberID, s)
	if err != nil {
		sugar.Error(err)
		return err
	}

	return err
}

func (router StripeRouter) WebHook(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sugar()
	sugar := router.logger.Sugar()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		sugar.Errorf("Error reading request body %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := webhook.ConstructEvent(body, req.Header.Get("Stripe-Signature"), router.signingKey)
	if err != nil {
		sugar.Errorf("Error verifying webhook signature: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sugar.Infof("Stripe event received: %s", event.Type)

	switch event.Type {

	// Occurs whenever a customer is signed up for a new plan.
	case "customer.subscription.created":
		s := stripeSdk.Subscription{}
		if err := json.Unmarshal(event.Data.Raw, &s); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sugar.Info(s)

		w.WriteHeader(http.StatusOK)

		go func() {
			err := router.onSubscription(&s)
			if err != nil {
				sugar.Error(err)
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
		sugar.Info(s)
		w.WriteHeader(http.StatusOK)

		go func() {
			err := router.onSubscription(&s)
			if err != nil {
				sugar.Error(err)
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
