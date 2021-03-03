package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
)

type PaymentIntent struct {
	ID                 string                                 `json:"id"`
	Amount             float64                                `json:"amount"`
	CanceledAtUTC      chrono.Time                            `json:"canceledAtUtc"`
	CancellationReason stripe.PaymentIntentCancellationReason `json:"cancellationReason"`
	ClientSecret       null.String                            `json:"clientSecret"`
	CreatedUtc         chrono.Time                            `json:"createdUtc"`
	Currency           string                                 `json:"currency"`
	CustomerID         string                                 `json:"customerId"`
	InvoiceID          string                                 `json:"invoiceId"`
	LiveMode           bool                                   `json:"liveMode"`
	PaymentMethodID    string                                 `json:"paymentMethodId"`
	// requires_payment_method,
	// requires_confirmation,
	// requires_action,
	// processing,
	// requires_capture,
	// canceled,
	// succeeded
	// See https://stripe.com/docs/payments/intents#intent-statuses
	Status stripe.PaymentIntentStatus `json:"status"`
}

func (pi PaymentIntent) RequiresAction() bool {
	return pi.Status == stripe.PaymentIntentStatusRequiresAction
}

// IsZero tests whether payment intent is missing from the response.
func (pi PaymentIntent) IsZero() bool {
	return pi.ID == ""
}

func NewPaymentIntent(pi *stripe.PaymentIntent) PaymentIntent {
	if pi == nil {
		return PaymentIntent{}
	}

	return PaymentIntent{
		ID:                 pi.ID,
		Amount:             float64(pi.Amount) / 100,
		CanceledAtUTC:      chrono.TimeFrom(dt.FromUnix(pi.CanceledAt)),
		CancellationReason: pi.CancellationReason,
		ClientSecret:       null.NewString(pi.ClientSecret, pi.ClientSecret != ""),
		CreatedUtc:         chrono.TimeFrom(dt.FromUnix(pi.Created)),
		Currency:           pi.Currency,
		CustomerID:         pi.Customer.ID,
		InvoiceID:          pi.Invoice.ID,
		LiveMode:           pi.Livemode,
		PaymentMethodID:    pi.PaymentMethod.ID,
		Status:             pi.Status,
	}
}
