package striperepo

import (
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/stripe/stripe-go/v72"
)

// NewSetupCheckout
// A SetupIntent guides you through the process of setting up and saving a customer's payment credentials for future payment.
func (c Client) NewSetupCheckout(cusID string) (*stripe.SetupIntent, error) {
	params := &stripe.SetupIntentParams{
		Customer: stripe.String(cusID),
		Usage:    stripe.String("on_session"),
	}

	return c.sc.SetupIntents.New(params)
}

func (c Client) NewCheckoutSession(p ftcStripe.CheckoutParams) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(p.Input.SuccessURL),
		CancelURL:  stripe.String(p.Input.CancelURL),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(p.Plan.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
	}

	if p.Account.StripeID.Valid {
		params.Customer = stripe.String(p.Account.StripeID.String)
	} else {
		params.CustomerEmail = stripe.String(p.Account.Email)
	}

	return c.sc.CheckoutSessions.New(params)
}
