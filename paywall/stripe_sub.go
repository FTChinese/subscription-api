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

type PaymentOutcome int

const (
	OutcomeUnknown PaymentOutcome = iota
	OutcomeSuccess
	OutcomeFailure
	OutcomeRequiresAction
)

type StripeSub struct {
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

// Bridge between chrono pkg and unix timestamp.
func canonicalizeUnix(s int64) time.Time {
	if s > 0 {
		return time.Unix(s, 0)
	}

	return time.Time{}
}

func NewStripeSub(s *stripe.Subscription) StripeSub {
	return StripeSub{
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		Created:            chrono.TimeFrom(canonicalizeUnix(s.Created)),
		CurrentPeriodEnd:   chrono.TimeFrom(canonicalizeUnix(s.CurrentPeriodEnd)),
		CurrentPeriodStart: chrono.TimeFrom(canonicalizeUnix(s.CurrentPeriodStart)),
		EndedAt:            chrono.TimeFrom(canonicalizeUnix(s.EndedAt)),
		LatestInvoice:      NewStripeInvoice(s.LatestInvoice),
		StartDate:          chrono.TimeFrom(canonicalizeUnix(s.StartDate)),
		Status:             s.Status,
	}
}

func (s StripeSub) Outcome() PaymentOutcome {
	if s.Status == stripe.SubscriptionStatusActive && s.LatestInvoice.PaymentIntent.Status == stripe.PaymentIntentStatusSucceeded {
		return OutcomeSuccess
	}

	if s.Status == stripe.SubscriptionStatusIncomplete && s.LatestInvoice.PaymentIntent.Status == stripe.PaymentIntentStatusRequiresPaymentMethod {
		return OutcomeFailure
	}

	if s.Status == stripe.SubscriptionStatusIncomplete && s.LatestInvoice.PaymentIntent.Status == stripe.PaymentIntentStatusRequiresAction {
		return OutcomeRequiresAction
	}

	return OutcomeUnknown
}

type StripeInvoice struct {
	Created          chrono.Time     `json:"created"`
	Currency         stripe.Currency `json:"currency"`
	HostedInvoiceURL string          `json:"hostedInvoiceUrl"`
	InvoicePDF       string          `json:"invoicePdf"`
	// A unique, identifying string that appears on emails sent to the customer for this invoice.
	Number string `json:"number"`
	Paid   bool   `json:"paid"`

	PaymentIntent PaymentIntent `json:"paymentIntent"`
	// This is the transaction number that appears on email receipts sent for this invoice.
	ReceiptNumber string `json:"receiptNumber"`
}

func NewStripeInvoice(i *stripe.Invoice) StripeInvoice {
	return StripeInvoice{
		Created:          chrono.TimeFrom(canonicalizeUnix(i.Created)),
		Currency:         i.Currency,
		HostedInvoiceURL: i.HostedInvoiceURL,
		InvoicePDF:       i.InvoicePDF,
		Number:           i.Number,
		Paid:             i.Paid,
		PaymentIntent:    NewPaymentIntent(i.PaymentIntent),
		ReceiptNumber:    i.ReceiptNumber,
	}
}

type PaymentIntent struct {
	ClientSecret string `json:"clientSecret"`
	// Status of this PaymentIntent, one of requires_payment_method, requires_confirmation, requires_action, processing, requires_capture, canceled, or succeeded
	Status stripe.PaymentIntentStatus `json:"status"`
}

func NewPaymentIntent(pi *stripe.PaymentIntent) PaymentIntent {
	return PaymentIntent{
		ClientSecret: pi.ClientSecret,
		Status:       pi.Status,
	}
}
