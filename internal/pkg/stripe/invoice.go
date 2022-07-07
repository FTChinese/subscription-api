package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/lib/sq"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
)

type Invoice struct {
	IsFromStripe         bool                    `json:"-"`
	ID                   string                  `json:"id" db:"id"`
	AutoAdvance          bool                    `json:"autoAdvance" db:"auto_advance"`
	ChargeID             string                  `json:"chargeId" db:"charge_id"`
	CollectionMethod     InvoiceCollectionMethod `json:"collectionMethod" db:"collection_method"`
	Currency             string                  `json:"currency" db:"currency"`
	CustomerID           string                  `json:"customerId" db:"customer_id"`
	DefaultPaymentMethod null.String             `json:"defaultPaymentMethod" db:"default_payment_method"`
	Discounts            sq.StringList           `json:"discounts" db:"discount_ids"`
	HostedInvoiceURL     null.String             `json:"hostedInvoiceUrl" db:"hosted_invoice_url"`
	LiveMode             bool                    `json:"liveMode" db:"live_mode"`
	Paid                 bool                    `json:"paid" db:"paid"`
	PaymentIntentID      string                  `json:"paymentIntentId" db:"payment_intent_id"`
	PeriodEndUTC         chrono.Time             `json:"periodEndUtc" db:"period_end_utc"`
	PeriodStartUTC       chrono.Time             `json:"periodStartUtc" db:"period_start_utc"`
	ReceiptNumber        string                  `json:"receiptNumber" db:"receipt_number"`
	Status               InvoiceStatus           `json:"status" db:"invoice_status"`
	SubscriptionID       null.String             `json:"subscriptionId" db:"subscription_id"`
	Total                int64                   `json:"total" db:"total"`
	Created              int64                   `json:"created" db:"created"`
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
		IsFromStripe:         true,
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
