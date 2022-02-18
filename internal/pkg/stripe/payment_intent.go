package stripe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/subscription-api/lib/collection"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
)

type NextActionJSON struct {
	*stripe.PaymentIntentNextAction
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (na NextActionJSON) Value() (driver.Value, error) {
	if na.PaymentIntentNextAction == nil {
		return nil, nil
	}

	b, err := json.Marshal(na.PaymentIntentNextAction)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (na *NextActionJSON) Scan(src interface{}) error {
	if src == nil {
		*na = NextActionJSON{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp stripe.PaymentIntentNextAction
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*na = NextActionJSON{&tmp}
		return nil

	default:
		return errors.New("incompatible type to scan to PaymentIntent")
	}
}

type PaymentIntent struct {
	ID                 string                                 `json:"id" db:"id"`
	Amount             int64                                  `json:"amount" db:"amount"`
	AmountReceived     int64                                  `json:"-" db:"amount_received"`
	CanceledAt         int64                                  `json:"-" db:"canceled_at"`
	CancellationReason stripe.PaymentIntentCancellationReason `json:"cancellationReason" db:"cancellation_reason"`
	ClientSecret       null.String                            `json:"clientSecret" db:"client_secret"`
	Created            int64                                  `json:"-" db:"created"`
	Currency           string                                 `json:"currency" db:"currency"`
	CustomerID         string                                 `json:"customerId" db:"customer_id"`
	InvoiceID          string                                 `json:"invoiceId" db:"invoice_id"`
	LiveMode           bool                                   `json:"liveMode" db:"live_mode"`
	NextAction         NextActionJSON                         `json:"next_action" db:"next_action"`
	PaymentMethodID    string                                 `json:"paymentMethodId" db:"payment_method_id"`
	PaymentMethodTypes collection.StringList                  `json:"-" db:"payment_method_types"`
	ReceiptEmail       string                                 `json:"-" db:"receipt_email"`
	SetupFutureUsage   stripe.PaymentIntentSetupFutureUsage   `json:"-" db:"setup_future_usage"`
	// requires_payment_method,
	// requires_confirmation,
	// requires_action,
	// processing,
	// requires_capture,
	// canceled,
	// succeeded
	// See https://stripe.com/docs/payments/intents#intent-statuses
	Status stripe.PaymentIntentStatus `json:"status" db:"intent_status"`
}

// NewPaymentIntent transforms stripe's payment intent.
// The PaymentIntent is generated when the invoice is finalized, and can then be used to pay the invoice.
// Trialing might not have one.
func NewPaymentIntent(pi *stripe.PaymentIntent) PaymentIntent {
	if pi == nil {
		return PaymentIntent{}
	}

	var cusID string
	if pi.Customer != nil {
		cusID = pi.Customer.ID
	}

	var invID string
	if pi.Invoice != nil {
		invID = pi.Invoice.ID
	}

	var pmID string
	if pi.PaymentMethod != nil {
		pmID = pi.PaymentMethod.ID
	}

	return PaymentIntent{
		ID:                 pi.ID,
		Amount:             pi.Amount,
		AmountReceived:     pi.AmountReceived,
		CanceledAt:         pi.CanceledAt,
		CancellationReason: pi.CancellationReason,
		ClientSecret:       null.NewString(pi.ClientSecret, pi.ClientSecret != ""),
		Created:            pi.Created,
		Currency:           pi.Currency,
		CustomerID:         cusID,
		InvoiceID:          invID,
		LiveMode:           pi.Livemode,
		NextAction:         NextActionJSON{pi.NextAction},
		PaymentMethodID:    pmID,
		PaymentMethodTypes: pi.PaymentMethodTypes,
		ReceiptEmail:       pi.ReceiptEmail,
		SetupFutureUsage:   pi.SetupFutureUsage,
		Status:             pi.Status,
	}
}

func (pi PaymentIntent) RequiresAction() bool {
	return pi.Status == stripe.PaymentIntentStatusRequiresAction
}

// IsZero tests whether payment intent is missing from the response.
func (pi PaymentIntent) IsZero() bool {
	return pi.ID == ""
}
