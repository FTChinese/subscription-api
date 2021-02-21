package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
)

type Invoice struct {
	ID                   string      `db:"id"`
	AutoAdvance          bool        `db:"auto_advance"`
	ChargeID             string      `db:"charge_id"`
	CollectionMethod     string      `db:"collection_method"`
	Currency             string      `db:"currency"`
	CustomerID           string      `db:"customer_id"`
	DefaultPaymentMethod null.String `db:"default_payment_method"`
	HostedInvoiceURL     null.String `db:"hosted_invoice_url"`
	Paid                 bool        `db:"paid"`
	PaymentIntentID      string      `db:"payment_intent_id"`
	PeriodEndUTC         chrono.Time `db:"period_end_utc"`
	PeriodStartUTC       chrono.Time `db:"period_start_utc"`
	ReceiptNumber        string      `db:"receipt_number"`
	Status               string      `db:"invoice_status"`
	Total                int64       `db:"total"`
	CreatedUTC           chrono.Time `db:"created_utc"`
	UpdatedUTC           chrono.Time `db:"updated_utc"`
}

func NewInvoice(si *stripe.Invoice) Invoice {
	var cm string
	if si.CollectionMethod != nil {
		cm = string(*si.CollectionMethod)
	}

	var pmID string
	if si.DefaultPaymentMethod != nil {
		pmID = si.DefaultPaymentMethod.ID
	}

	var piID string
	if si.PaymentIntent != nil {
		piID = si.PaymentIntent.ID
	}

	return Invoice{
		ID:                   si.ID,
		AutoAdvance:          si.AutoAdvance,
		ChargeID:             si.Charge.ID,
		CollectionMethod:     cm,
		Currency:             string(si.Currency),
		CustomerID:           si.Customer.ID,
		DefaultPaymentMethod: null.NewString(pmID, pmID == ""),
		HostedInvoiceURL:     null.NewString(si.HostedInvoiceURL, si.HostedInvoiceURL != ""),
		Paid:                 si.Paid,
		PaymentIntentID:      piID,
		PeriodEndUTC:         chrono.TimeFrom(dt.FromUnix(si.PeriodEnd)),
		PeriodStartUTC:       chrono.TimeFrom(dt.FromUnix(si.PeriodStart)),
		ReceiptNumber:        si.ReceiptNumber,
		Status:               string(si.Status),
		Total:                si.Total,
		CreatedUTC:           chrono.TimeFrom(dt.FromUnix(si.Created)),
		UpdatedUTC:           chrono.TimeNow(),
	}
}
