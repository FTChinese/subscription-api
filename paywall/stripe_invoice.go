package paywall

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go"
	"strings"
)

type StripeInvoice struct {
	Created          chrono.Time     `json:"created"`
	Currency         stripe.Currency `json:"currency"`
	HostedInvoiceURL string          `json:"hostedInvoiceUrl"`
	ID               string          `json:"id"`
	InvoicePDF       string          `json:"invoicePdf"`
	Livemode         bool            `json:"-"`
	// A unique, identifying string that appears on emails sent to the customer for this invoice.
	Number        string        `json:"number"`
	Paid          bool          `json:"paid"`
	PaymentIntent PaymentIntent `json:"paymentIntent"`
	// This is the transaction number that appears on email receipts sent for this invoice.
	ReceiptNumber string `json:"receiptNumber"`
}

func NewStripeInvoice(i *stripe.Invoice) StripeInvoice {
	if i == nil {
		return StripeInvoice{}
	}

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
	ClientSecret string                          `json:"clientSecret"`
	ID           string                          `json:"id"`
	NextAction   *stripe.PaymentIntentNextAction `json:"nextAction,omitempty"`
	// Status of this PaymentIntent, one of requires_payment_method, requires_confirmation, requires_action, processing, requires_capture, canceled, or succeeded
	Status stripe.PaymentIntentStatus `json:"status"`
}

func NewPaymentIntent(pi *stripe.PaymentIntent) PaymentIntent {
	if pi == nil {
		return PaymentIntent{}
	}

	return PaymentIntent{
		ClientSecret: pi.ClientSecret,
		ID:           pi.ID,
		NextAction:   pi.NextAction,
		Status:       pi.Status,
	}
}

type EmailedInvoice struct {
	*stripe.Invoice
}

func (i EmailedInvoice) CreationTime() chrono.Time {
	return chrono.TimeFrom(canonicalizeUnix(i.Created))
}

func (i EmailedInvoice) BuildFtcPlan() (Plan, error) {
	if i.Lines == nil {
		return Plan{}, errors.New("empty lines")
	}

	if len(i.Lines.Data) == 0 {
		return Plan{}, errors.New("empty lines.data")
	}

	stripePlan := i.Lines.Data[0].Plan

	ftcPlan, err := GetStripeToFtcPlans(i.Livemode).FindPlan(stripePlan.ID)
	if err != nil {
		return Plan{}, err
	}

	return ftcPlan.WithStripe(*stripePlan), nil
}

func (i EmailedInvoice) Price() string {
	return fmt.Sprintf("%s%.2f", strings.ToUpper(string(i.Currency)), float64(i.Total/100))
}
