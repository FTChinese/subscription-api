package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/lib/sq"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
)

type Invoice struct {
	ID                   string                  `db:"id"`
	AutoAdvance          bool                    `db:"auto_advance"`
	ChargeID             string                  `db:"charge_id"`
	CollectionMethod     InvoiceCollectionMethod `db:"collection_method"`
	Currency             string                  `db:"currency"`
	CustomerID           string                  `db:"customer_id"`
	DefaultPaymentMethod null.String             `db:"default_payment_method"`
	Discounts            sq.StringList           `db:"discount_ids"`
	HostedInvoiceURL     null.String             `db:"hosted_invoice_url"`
	LiveMode             bool                    `db:"live_mode"`
	Paid                 bool                    `db:"paid"`
	PaymentIntentID      string                  `db:"payment_intent_id"`
	PeriodEndUTC         chrono.Time             `db:"period_end_utc"`
	PeriodStartUTC       chrono.Time             `db:"period_start_utc"`
	ReceiptNumber        string                  `db:"receipt_number"`
	Status               InvoiceStatus           `db:"invoice_status"`
	SubscriptionID       null.String             `db:"subscription_id"`
	Total                int64                   `db:"total"`
	Created              int64                   `db:"created"`
}

func NewInvoice(si *stripe.Invoice) Invoice {

	// Those fields are nil if you get it from subscripiton
	// without expansion.
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

	var chargeID string
	if si.Charge != nil {
		chargeID = si.Charge.ID
	}

	var cusID string
	if si.Customer != nil {
		cusID = si.Customer.ID
	}

	return Invoice{
		ID:                   si.ID,
		AutoAdvance:          si.AutoAdvance,
		ChargeID:             chargeID,
		CollectionMethod:     newInvoiceCollectionMethod(si.CollectionMethod),
		Currency:             string(si.Currency),
		CustomerID:           cusID,
		DefaultPaymentMethod: null.NewString(pmID, pmID == ""),
		Discounts:            collectDiscountIDs(si.Discounts),
		HostedInvoiceURL:     null.NewString(si.HostedInvoiceURL, si.HostedInvoiceURL != ""),
		LiveMode:             si.Livemode,
		Paid:                 si.Paid,
		PaymentIntentID:      piID,
		PeriodEndUTC:         chrono.TimeFrom(dt.FromUnix(si.PeriodEnd)),
		PeriodStartUTC:       chrono.TimeFrom(dt.FromUnix(si.PeriodStart)),
		ReceiptNumber:        si.ReceiptNumber,
		Status:               InvoiceStatus{si.Status},
		SubscriptionID:       null.NewString(subsID, subsID != ""),
		Total:                si.Total,
		Created:              si.Created,
	}
}

func collectDiscountIDs(discounts []*stripe.Discount) []string {
	var list = make([]string, 0)

	for _, v := range discounts {
		list = append(list, v.ID)
	}

	return list
}
