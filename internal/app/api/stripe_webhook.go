package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	sdk "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/webhook"
	"io/ioutil"
	"net/http"
)

// WebHook to listen to those events:
// - customer.updated: Occurs whenever any property of a customer changes.
// - customer.subscription.created
// - customer.subscription.updated
// - customer.subscription.deleted
// - invoice.created
// - invoice.finalized
// - invoice.payment_action_required
// - invoice.payment_failed
// - invoice.payment_succeeded
// - invoice.upcoming
// See https://stripe.com/docs/api/events/types
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
	case "customer.created", "customer.updated":
		var rawCus sdk.Customer
		err := json.Unmarshal(event.Data.Raw, &rawCus)
		if err != nil {
			sugar.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		go func() {
			_ = router.eventCustomer(rawCus)
		}()
		w.WriteHeader(http.StatusOK)

		// create occurs whenever a customer is signed up for a new plan.
	// update occurs whenever a subscription changes (e.g., switching from one plan to another, or changing the status from trial to active).
	case "customer.subscription.created",
		"customer.subscription.updated",
		"customer.subscription.deleted":
		s := sdk.Subscription{}
		if err := json.Unmarshal(event.Data.Raw, &s); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		go func() {
			err := router.eventSubscription(&s)
			if err != nil {
				sugar.Error(err)
			}
		}()
		w.WriteHeader(http.StatusOK)

	case "setup_intent.succeeded":
		si := sdk.SetupIntent{}
		if err := json.Unmarshal(event.Data.Raw, &si); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		go func() {
			_ = router.eventSetupIntent(si)
		}()
		w.WriteHeader(http.StatusOK)

	case "setup_intent.canceled":
		si := sdk.SetupIntent{}
		if err := json.Unmarshal(event.Data.Raw, &si); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		go func() {
			err := router.Env.UpsertSetupIntent(stripe.NewSetupIntent(&si))
			if err != nil {
				sugar.Error(err)
			}
		}()
		w.WriteHeader(http.StatusOK)

	// A few days prior to renewal, your site receives an invoice.upcoming event at the webhook endpoint.
	case "invoice.created",
		"invoice.payment_failed",
		"invoice.payment_action_required",
		"invoice.upcoming",
		"invoice.finalized":
		// Stripe waits an hour after receiving a successful response to the invoice.created event before attempting payment.
		// If a successful response isn’t received within 72 hours, Stripe attempts to finalize and send the invoice.
		// In live mode, if your webhook endpoint does not respond properly, Stripe continues retrying the webhook notification for up to three days with an exponential back off
		var i sdk.Invoice
		if err := json.Unmarshal(event.Data.Raw, &i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sugar.Infof("invoice: %v", i)
		go func() {
			err := router.Env.UpsertInvoice(stripe.NewInvoice(&i))
			if err != nil {
				sugar.Error(err)
			}
		}()
		w.WriteHeader(http.StatusOK)

	// Set default payment method after payment succeeded.
	// Retrieve the payment intent by invoice.payment_intent.
	// Then set the payment intent's payment method id to subscription.
	// See https://stripe.com/docs/billing/subscriptions/build-subscription?ui=elements#default-payment-method
	case "invoice.payment_succeeded":
		var invoice sdk.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			sugar.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		go func() {
			_ = router.eventPaymentSucceeded(invoice)
		}()
		w.WriteHeader(http.StatusOK)

	case "payment_method.attached",
		"payment_method.automatically_updated",
		"payment_method.updated":

		var rawPM sdk.PaymentMethod
		err := json.Unmarshal(event.Data.Raw, &rawPM)
		if err != nil {
			sugar.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		go func() {
			_ = router.eventPaymentMethod(rawPM)
		}()
		w.WriteHeader(http.StatusOK)

	default:
		w.WriteHeader(http.StatusOK)
	}
}

// eventCustomer handles Stripe webhook events:
// - customer.created
// - customer.updated
// If the default payment method is set on a customer, we will try to
// duplicate the payment method data from Stripe if it is not yet saved
// in our db yet.
func (router StripeRouter) eventCustomer(rawCus sdk.Customer) error {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	baseAccount, err := router.ReaderRepo.BaseAccountByStripeID(rawCus.ID)
	if err != nil {
		sugar.Error(err)
		return err
	}

	cus := stripe.NewCustomer(baseAccount.FtcID, &rawCus)

	err = router.Env.UpsertCustomer(cus)
	if err != nil {
		sugar.Error(err)
		return err
	}

	if cus.DefaultPaymentMethodID.IsZero() {
		return nil
	}

	// Next checks if payment method saved in our db.
	pm, err := router.Env.LoadOrFetchPaymentMethod(cus.DefaultPaymentMethodID.String, false)
	if err != nil {
		return nil
	}

	if !pm.IsFromStripe {
		return nil
	}

	err = router.Env.UpsertPaymentMethod(pm)
	if err != nil {
		return err
	}

	return nil
}

func (router StripeRouter) eventSetupIntent(rawSI sdk.SetupIntent) error {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	si := stripe.NewSetupIntent(&rawSI)
	err := router.Env.UpsertSetupIntent(si)
	if err != nil {
		sugar.Error(err)
		return err
	}

	if si.PaymentMethodID.IsZero() {
		return nil
	}

	pm, err := router.Env.LoadOrFetchPaymentMethod(si.PaymentMethodID.String, true)
	if err != nil {
		sugar.Error(err)
		return err
	}

	// Save/Update payment method
	err = router.Env.UpsertPaymentMethod(pm)
	if err != nil {
		sugar.Error(err)
		return err
	}

	_, err = router.Env.SetCusDefaultPaymentIfMissing(si.CustomerID, si.PaymentMethodID.String)
	if err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

// eventPaymentMethod handles payment method event.
func (router StripeRouter) eventPaymentMethod(rawPM sdk.PaymentMethod) error {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	pm := stripe.NewPaymentMethod(&rawPM)
	err := router.Env.UpsertPaymentMethod(pm)
	if err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

// eventPaymentSucceeded sets payment method extracted from
// invoice's payment intent to the subscription contained
// in this invoice.
func (router StripeRouter) eventPaymentSucceeded(rawInvoice sdk.Invoice) error {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	pi, err := router.Env.Client.FetchPaymentIntent(
		rawInvoice.PaymentIntent.ID)

	if err != nil {
		sugar.Error(err)
		return err
	}

	// Set default payment method on subscription
	subs, err := router.Env.Client.SetSubsDefaultPaymentMethod(
		rawInvoice.Subscription.ID,
		pi.PaymentMethod.ID,
	)
	if err != nil {
		sugar.Error(err)
		return err
	}

	sugar.Infof("Default payment method set for subscription: %s - %s", subs.ID, pi.PaymentMethod.ID)

	pm, err := router.Env.LoadOrFetchPaymentMethod(pi.PaymentMethod.ID, false)
	if err != nil {
		sugar.Error(err)
		return err
	}

	if !pm.IsFromStripe {
		return nil
	}

	// Save/Update payment method
	err = router.Env.UpsertPaymentMethod(pm)
	if err != nil {
		sugar.Error(err)
		return err
	}

	// Save the invoice.
	err = router.Env.UpsertInvoice(stripe.NewInvoice(&rawInvoice))
	if err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

// Handle subscription received by webhook.
func (router StripeRouter) eventSubscription(ss *sdk.Subscription) error {

	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	// Find user account by stripe customer id.
	account, err := router.ReaderRepo.BaseAccountByStripeID(ss.Customer.ID)
	if err != nil {
		sugar.Error(err)
		// If user account is not found,
		// we still want to save this subscription.
		// Stop here since we don't know who's using this subscription.
		if err != sql.ErrNoRows {
			return err
		}

		err = router.Env.UpsertSubs(
			stripe.NewSubs("", ss),
			false)
		if err != nil {
			sugar.Error(err)
		}

		return err
	}

	// stripe.Subs could always be created regardless of user account present or not.
	userIDs := account.CompoundIDs()
	subs := stripe.NewSubs("", ss)

	result, err := router.Env.OnWebhookSubs(subs, userIDs)
	if err != nil {
		sugar.Error(err)

		var whe stripe.WebhookError
		if errors.As(err, &whe) {
			err := router.Env.SaveWebhookError(whe)
			if err != nil {
				sugar.Error(err)
			}
		}

		return err
	}

	// Update subscription
	err = router.Env.UpsertSubs(
		stripe.NewSubs("", ss), false)

	if err != nil {
		sugar.Error(err)
	}

	// Backup old membership
	err = router.ReaderRepo.VersionMembership(result.Versioned)

	if err != nil {
		sugar.Error(err)
	}

	return err
}
