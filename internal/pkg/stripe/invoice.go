package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
)

type Invoice struct {
	ID                   string                         `db:"id"`
	AutoAdvance          bool                           `db:"auto_advance"`
	ChargeID             string                         `db:"charge_id"`
	CollectionMethod     stripe.InvoiceCollectionMethod `db:"collection_method"`
	Currency             string                         `db:"currency"`
	CustomerID           string                         `db:"customer_id"`
	DefaultPaymentMethod null.String                    `db:"default_payment_method"`
	HostedInvoiceURL     null.String                    `db:"hosted_invoice_url"`
	LiveMode             bool                           `db:"live_mode"`
	Paid                 bool                           `db:"paid"`
	PaymentIntentID      string                         `db:"payment_intent_id"`
	PeriodEndUTC         chrono.Time                    `db:"period_end_utc"`
	PeriodStartUTC       chrono.Time                    `db:"period_start_utc"`
	ReceiptNumber        string                         `db:"receipt_number"`
	Status               stripe.InvoiceStatus           `db:"invoice_status"`
	SubscriptionID       null.String                    `db:"subscription_id"`
	Total                int64                          `db:"total"`
	Created              int64                          `db:"created"`
}

func NewInvoice(si *stripe.Invoice) Invoice {

	var pmID string
	if si.DefaultPaymentMethod != nil {
		pmID = si.DefaultPaymentMethod.ID
	}

	var subsID string
	if si.Subscription != nil {
		subsID = si.Subscription.ID
	}

	var piID string
	if si.PaymentIntent != nil {
		piID = si.PaymentIntent.ID
	}

	return Invoice{
		ID:                   si.ID,
		AutoAdvance:          si.AutoAdvance,
		ChargeID:             si.Charge.ID,
		CollectionMethod:     *si.CollectionMethod,
		Currency:             string(si.Currency),
		CustomerID:           si.Customer.ID,
		DefaultPaymentMethod: null.NewString(pmID, pmID == ""),
		HostedInvoiceURL:     null.NewString(si.HostedInvoiceURL, si.HostedInvoiceURL != ""),
		LiveMode:             si.Livemode,
		Paid:                 si.Paid,
		PaymentIntentID:      piID,
		PeriodEndUTC:         chrono.TimeFrom(dt.FromUnix(si.PeriodEnd)),
		PeriodStartUTC:       chrono.TimeFrom(dt.FromUnix(si.PeriodStart)),
		ReceiptNumber:        si.ReceiptNumber,
		Status:               si.Status,
		SubscriptionID:       null.NewString(subsID, subsID != ""),
		Total:                si.Total,
		Created:              si.Created,
	}
}
