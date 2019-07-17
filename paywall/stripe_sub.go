package paywall

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
	"time"
)

type StripeSubParams struct {
	Tier                 enum.Tier   `json:"tier"`
	Cycle                enum.Cycle  `json:"cycle"`
	Customer             string      `json:"customer"`
	Coupon               null.String `json:"coupon"`
	DefaultPaymentMethod null.String `json:"defaultPaymentMethod"`
	PlanID               string      `json:"-"`
	IdempotencyKey       string      `json:"idempotency"`
}

func (s StripeSubParams) Key() string {
	return s.Tier.String() + "_" + s.Cycle.String()
}

func extractStripePlanID(s *stripe.Subscription) (string, error) {
	if s.Items == nil {
		return "", errors.New("stripe subscription items are nil")
	}
	if len(s.Items.Data) == 0 {
		return "", errors.New("stripe subscription items are empty")
	}

	if s.Items.Data[0].Plan == nil {
		return "", errors.New("stripe subscription plan is nil")
	}

	return s.Items.Data[0].Plan.ID, nil
}

type StripeSub struct {
	// A date in the future at which the subscription will automatically get canceled. Nullable.
	CancelAt           chrono.Time   `json:"cancelAt"`
	CancelAtPeriodEnd  bool          `json:"cancelAtPeriodEnd"`
	Created            chrono.Time   `json:"created"`
	CurrentPeriodEnd   chrono.Time   `json:"currentPeriodEnd"`
	CurrentPeriodStart chrono.Time   `json:"currentPeriodStart"`
	EndedAt            chrono.Time   `json:"endedAt"`
	LatestInvoice      StripeInvoice `json:"latestInvoice"`
	StartDate          chrono.Time   `json:"startDate"`

	// Possible values are incomplete, incomplete_expired, trialing, active, past_due, canceled, or unpaid.
	Status stripe.SubscriptionStatus `json:"status"`
}

func NewStripeSub(s *stripe.Subscription) StripeSub {
	return StripeSub{
		CancelAt:           chrono.TimeFrom(time.Unix(s.CancelAt, 0)),
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		Created:            chrono.TimeFrom(time.Unix(s.Created, 0)),
		CurrentPeriodEnd:   chrono.TimeFrom(time.Unix(s.CurrentPeriodEnd, 0)),
		CurrentPeriodStart: chrono.TimeFrom(time.Unix(s.CurrentPeriodStart, 0)),
		EndedAt:            chrono.TimeFrom(time.Unix(s.EndedAt, 0)),
		LatestInvoice:      NewStripeInvoice(s.LatestInvoice),
		StartDate:          chrono.TimeFrom(time.Unix(s.StartDate, 0)),
		Status:             s.Status,
	}
}

type StripeInvoice struct {
	Created          chrono.Time     `json:"created"`
	Currency         stripe.Currency `json:"currency"`
	HostedInvoiceURL string          `json:"hostedInvoiceUrl"`
	InvoicePDF       string          `json:"invoicePdf"`
	// A unique, identifying string that appears on emails sent to the customer for this invoice.
	Number string `json:"number"`
	Paid   bool   `json:"paid"`
	// Status of this PaymentIntent, one of requires_payment_method, requires_confirmation, requires_action, processing, requires_capture, canceled, or succeeded
	PaymentIntentStatus stripe.PaymentIntentStatus `json:"paymentIntentStatus"`
	// This is the transaction number that appears on email receipts sent for this invoice.
	ReceiptNumber string `json:"receiptNumber"`
	// The status of the invoice, one of draft, open, paid, uncollectible, or void
	Status stripe.InvoiceBillingStatus `json:"status"`
}

func NewStripeInvoice(i *stripe.Invoice) StripeInvoice {
	return StripeInvoice{
		Created:             chrono.TimeFrom(time.Unix(i.Created, 0)),
		Currency:            i.Currency,
		HostedInvoiceURL:    i.HostedInvoiceURL,
		InvoicePDF:          i.InvoicePDF,
		Number:              i.Number,
		Paid:                i.Paid,
		PaymentIntentStatus: i.PaymentIntent.Status,
		ReceiptNumber:       i.ReceiptNumber,
		Status:              i.Status,
	}
}
