package stripe

import (
	"errors"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v71"
)

// PaymentResult tells client whether user a subscription is created successfully.
// If the payment requires further action, for example credit card authentication
// on the client side, the PaymentIntentClientSecret is populated.
type PaymentResult struct {
	RequiresAction            bool        `json:"requiresAction"`
	PaymentIntentClientSecret null.String `json:"paymentIntentClientSecret"`
}

// NewPaymentResult returns the a subscription's payment intent result to client.
func NewPaymentResult(ss *stripe.Subscription) (PaymentResult, error) {
	if ss.LatestInvoice == nil {
		return PaymentResult{}, errors.New("latest_invoice not expanded")
	}

	pi := ss.LatestInvoice.PaymentIntent
	if pi == nil {
		return PaymentResult{}, errors.New("latest_invoice.payment_intent not found")
	}

	return PaymentResult{
		RequiresAction:            pi.Status == stripe.PaymentIntentStatusRequiresAction,
		PaymentIntentClientSecret: null.StringFrom(pi.ClientSecret),
	}, nil
}
