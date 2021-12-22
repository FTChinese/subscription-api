package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/ids"
	stripeSdk "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/webhook"
	"io/ioutil"
	"net/http"
)

// Handle subscription received by webhook.
func (router StripeRouter) onSubscription(ss *stripeSdk.Subscription) error {

	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	// Find user account by stripe customer id.
	account, err := router.ReaderRepo.BaseAccountByStripeID(ss.Customer.ID)
	if err != nil {
		sugar.Error(err)
		// If user account is not found,
		// we still want to save this subscription.
		// Stop here since we don't know who's using this subscription.
		if err == sql.ErrNoRows {
			subs, err := stripe.NewSubs(ss, ids.UserIDs{})
			if err != nil {
				return err
			}

			err = router.StripeRepo.UpsertSubs(subs)
			if err != nil {
				return err
			}
		}
		return err
	}

	// stripe.Subs could always be created regardless of user account present or not.
	userIDs := account.CompoundIDs()
	subs, err := stripe.NewSubs(ss, userIDs)
	if err != nil {
		return err
	}

	result, err := router.StripeRepo.OnWebhookSubs(subs, userIDs)
	if err != nil {
		sugar.Error(err)

		var whe stripe.WebhookError
		if errors.As(err, &whe) {
			err := router.StripeRepo.SaveWebhookError(whe)
			if err != nil {
				sugar.Error(err)
			}
		}

		return err
	}

	err = router.ReaderRepo.VersionMembership(result.Versioned)

	if err != nil {
		sugar.Error(err)
	}

	return err
}

// WebHook to listen to those events:
// - customer.subscription.created
// - customer.subscription.updated
// - customer.subscription.deleted
// - invoice.created
// - invoice.finalized
// - invoice.payment_action_required
// - invoice.payment_failed
// - invoice.payment_succeeded
// - invoice.upcoming
func (router StripeRouter) WebHook(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sugar()
	sugar := router.Logger.Sugar()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		sugar.Errorf("Error reading request body %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := webhook.ConstructEvent(body, req.Header.Get("Stripe-Signature"), router.SigningKey)
	if err != nil {
		sugar.Errorf("Error verifying webhook signature: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sugar.Infof("Stripe event received: %s", event.Type)
	sugar.Infof("Stripe event raw data: %s", event.Data.Raw)

	switch event.Type {

	// create occurs whenever a customer is signed up for a new plan.
	// update occurs whenever a subscription changes (e.g., switching from one plan to another, or changing the status from trial to active).
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		s := stripeSdk.Subscription{}
		if err := json.Unmarshal(event.Data.Raw, &s); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		go func() {
			err := router.onSubscription(&s)
			if err != nil {
				sugar.Error(err)
			}
		}()

		w.WriteHeader(http.StatusOK)
		return

		// A few days prior to renewal, your site receives an invoice.upcoming event at the webhook endpoint.
	case "invoice.created", "invoice.payment_failed", "invoice.payment_action_required", "invoice.upcoming", "invoice.payment_succeeded", "invoice.finalized":
		// Stripe waits an hour after receiving a successful response to the invoice.created event before attempting payment.
		// If a successful response isn’t received within 72 hours, Stripe attempts to finalize and send the invoice.
		// In live mode, if your webhook endpoint does not respond properly, Stripe continues retrying the webhook notification for up to three days with an exponential back off
		var i stripeSdk.Invoice
		if err := json.Unmarshal(event.Data.Raw, &i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sugar.Infof("invoice.created: %v", i)

		go func() {
			err := router.StripeRepo.UpsertInvoice(stripe.NewInvoice(&i))
			if err != nil {
				sugar.Error(err)
			}
		}()

		w.WriteHeader(http.StatusOK)
		return

	default:
		w.WriteHeader(http.StatusOK)
	}
}
