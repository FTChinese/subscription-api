package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	sdk "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/webhook"
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
func (routes StripeRoutes) WebHook(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sugar()
	sugar := routes.logger.Sugar()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		sugar.Errorf("Error reading request body %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := webhook.ConstructEvent(body, req.Header.Get("Stripe-Signature"), routes.signingKey)
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
			_ = routes.eventCustomer(rawCus)
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
			err := routes.eventSubscription(&s)
			if err != nil {
				sugar.Error(err)
			}
		}()
		w.WriteHeader(http.StatusOK)

	case "coupon.created", "coupon.updated", "coupon.deleted":
		c := sdk.Coupon{}
		if err := json.Unmarshal(event.Data.Raw, &c); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		go func() {
			_ = routes.eventCoupon(c)
		}()

		w.WriteHeader(http.StatusOK)

	case "setup_intent.succeeded":
		si := sdk.SetupIntent{}
		if err := json.Unmarshal(event.Data.Raw, &si); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		go func() {
			_ = routes.eventSetupIntent(si)
		}()
		w.WriteHeader(http.StatusOK)

	case "setup_intent.canceled":
		si := sdk.SetupIntent{}
		if err := json.Unmarshal(event.Data.Raw, &si); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		go func() {
			err := routes.stripeRepo.UpsertSetupIntent(stripe.NewSetupIntent(&si))
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
			err := routes.stripeRepo.UpsertInvoice(stripe.NewInvoice(&i))
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
			_ = routes.eventPaymentSucceeded(invoice)
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
			_ = routes.eventPaymentMethod(rawPM)
		}()
		w.WriteHeader(http.StatusOK)

	case "price.created", "price.deleted", "price.updated":
		var rawPrice sdk.Price
		err := json.Unmarshal(event.Data.Raw, &rawPrice)
		if err != nil {
			sugar.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		go func() {
			_ = routes.eventPrice(rawPrice)
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
func (routes StripeRoutes) eventCustomer(rawCus sdk.Customer) error {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	baseAccount, err := routes.readerRepo.BaseAccountByStripeID(rawCus.ID)
	if err != nil {
		sugar.Error(err)
		return err
	}

	cus := stripe.NewCustomer(baseAccount.FtcID, &rawCus)

	err = routes.stripeRepo.UpsertCustomer(cus)
	if err != nil {
		sugar.Error(err)
		return err
	}

	if cus.DefaultPaymentMethodID.IsZero() {
		return nil
	}

	// Next checks if payment method saved in our db.
	pm, err := routes.stripeRepo.LoadOrFetchPaymentMethod(cus.DefaultPaymentMethodID.String, false)
	if err != nil {
		return nil
	}

	if !pm.IsFromStripe {
		return nil
	}

	err = routes.stripeRepo.UpsertPaymentMethod(pm)
	if err != nil {
		return err
	}

	return nil
}

func (routes StripeRoutes) eventCoupon(rawCoupon sdk.Coupon) error {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	coupon := price.NewStripeCoupon(&rawCoupon)

	err := routes.stripeRepo.UpsertCoupon(coupon)
	if err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

func (routes StripeRoutes) eventSetupIntent(rawSI sdk.SetupIntent) error {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	si := stripe.NewSetupIntent(&rawSI)
	err := routes.stripeRepo.UpsertSetupIntent(si)
	if err != nil {
		sugar.Error(err)
		return err
	}

	if si.PaymentMethodID.IsZero() {
		return nil
	}

	pm, err := routes.stripeRepo.LoadOrFetchPaymentMethod(si.PaymentMethodID.String, true)
	if err != nil {
		sugar.Error(err)
		return err
	}

	// Save/Update payment method
	err = routes.stripeRepo.UpsertPaymentMethod(pm)
	if err != nil {
		sugar.Error(err)
		return err
	}

	_, err = routes.stripeRepo.SetCusDefaultPaymentIfMissing(si.CustomerID, si.PaymentMethodID.String)
	if err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

// eventPaymentMethod handles payment method event.
func (routes StripeRoutes) eventPaymentMethod(rawPM sdk.PaymentMethod) error {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	pm := stripe.NewPaymentMethod(&rawPM)
	err := routes.stripeRepo.UpsertPaymentMethod(pm)
	if err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

// eventPaymentSucceeded sets payment method extracted from
// invoice's payment intent to the subscription contained
// in this invoice.
func (routes StripeRoutes) eventPaymentSucceeded(rawInvoice sdk.Invoice) error {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	pi, err := routes.stripeRepo.Client.FetchPaymentIntent(
		rawInvoice.PaymentIntent.ID)

	if err != nil {
		sugar.Error(err)
		return err
	}

	// Set default payment method on subscription
	subs, err := routes.stripeRepo.Client.SetSubsDefaultPaymentMethod(
		rawInvoice.Subscription.ID,
		pi.PaymentMethod.ID,
	)
	if err != nil {
		sugar.Error(err)
		return err
	}

	sugar.Infof("Default payment method set for subscription: %s - %s", subs.ID, pi.PaymentMethod.ID)

	pm, err := routes.stripeRepo.LoadOrFetchPaymentMethod(pi.PaymentMethod.ID, false)
	if err != nil {
		sugar.Error(err)
		return err
	}

	if !pm.IsFromStripe {
		return nil
	}

	// Save/Update payment method
	err = routes.stripeRepo.UpsertPaymentMethod(pm)
	if err != nil {
		sugar.Error(err)
		return err
	}

	// Save the invoice.
	err = routes.stripeRepo.UpsertInvoice(stripe.NewInvoice(&rawInvoice))
	if err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

func (routes StripeRoutes) eventPrice(rawPrice sdk.Price) error {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	p := price.NewStripePrice(&rawPrice)

	err := routes.stripeRepo.UpsertPrice(p)
	if err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

// Handle subscription received by webhook.
func (routes StripeRoutes) eventSubscription(ss *sdk.Subscription) error {

	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	// Find user account by stripe customer id.
	account, err := routes.readerRepo.BaseAccountByStripeID(ss.Customer.ID)
	if err != nil {
		sugar.Error(err)
		// If user account is not found,
		// we still want to save this subscription.
		// Stop here since we don't know who's using this subscription.
		if err != sql.ErrNoRows {
			return err
		}

		err = routes.stripeRepo.UpsertSubs(
			stripe.NewSubs("", ss),
			false)
		if err != nil {
			sugar.Error(err)
		}

		return err
	}

	// stripe.Subs could always be created regardless of user account present or not.
	subs := stripe.NewSubs("", ss)

	result, err := routes.stripeRepo.SyncSubs(
		account.CompoundIDs(),
		subs,
		reader.NewArchiver().ByStripe().ActionWebhook())
	if err != nil {
		sugar.Error(err)

		var whe stripe.WebhookError
		if errors.As(err, &whe) {
			err := routes.stripeRepo.SaveWebhookError(whe)
			if err != nil {
				sugar.Error(err)
			}
		}

		return err
	}

	routes.handleSubsResult(result)

	return nil
}
