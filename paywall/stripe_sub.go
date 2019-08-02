package paywall

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/stripe/stripe-go"
	"strings"
	"time"
)

// StripeSub is a reduced version of stripe.Subscription.
// Used as response when client asks for subscription data.
type StripeSub struct {
	AccountID AccountID `json:"-"`
	Coordinate
	CancelAtPeriodEnd  bool        `json:"cancelAtPeriodEnd"`
	Created            chrono.Time `json:"created"`
	CurrentPeriodEnd   chrono.Time `json:"currentPeriodEnd"`
	CurrentPeriodStart chrono.Time `json:"currentPeriodStart"`
	CustomerID         string      `json:"customerId"`
	EndedAt            chrono.Time `json:"endedAt"`
	ID                 string      `json:"id"`
	LatestInvoiceID    string      `json:"latestInvoiceId"`
	Livemode           bool        `json:"livemode"`
	StartDate          chrono.Time `json:"startDate"`

	// Possible values are incomplete, incomplete_expired, trialing, active, past_due, canceled, or unpaid.
	Status stripe.SubscriptionStatus `json:"status"`
}

// Bridge between chrono pkg and unix timestamp.
// Unix 0 represent year 1970, while Golang's zero time is really
// 0.
func canonicalizeUnix(s int64) time.Time {
	if s > 0 {
		return time.Unix(s, 0)
	}

	return time.Time{}
}

func NewStripeSub(s *stripe.Subscription) StripeSub {
	if s == nil {
		return StripeSub{}
	}

	return StripeSub{
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		Created:            chrono.TimeFrom(canonicalizeUnix(s.Created)),
		CurrentPeriodEnd:   chrono.TimeFrom(canonicalizeUnix(s.CurrentPeriodEnd)),
		CurrentPeriodStart: chrono.TimeFrom(canonicalizeUnix(s.CurrentPeriodStart)),
		CustomerID:         s.Customer.ID,
		EndedAt:            chrono.TimeFrom(canonicalizeUnix(s.EndedAt)),
		ID:                 s.ID,
		LatestInvoiceID:    s.LatestInvoice.ID,
		Livemode:           s.Livemode,
		StartDate:          chrono.TimeFrom(canonicalizeUnix(s.StartDate)),
		Status:             s.Status,
	}
}

func (s StripeSub) ReadableStatus() string {
	switch s.Status {
	case stripe.SubscriptionStatusActive:
		return "活跃"

	//  the initial payment attempt fails
	case stripe.SubscriptionStatusIncomplete:
		return "支付未完成，请在24小时内完成支付"

	// If the first invoice is not paid within 23 hours, the subscription transitions to incomplete_expired. This is a terminal state, the open invoice will be voided and no further invoices will be generated.
	case stripe.SubscriptionStatusIncompleteExpired:
		return "支付已过期"

	case stripe.SubscriptionStatusPastDue:
		// payment to renew it fails
		return "续费失败"

	case stripe.SubscriptionStatusCanceled:
		// when Stripe has exhausted all payment retry attempts.
		return "Stripe未能找到合适的支付方式，支付已取消"

	case stripe.SubscriptionStatusUnpaid:
		// when Stripe has exhausted all payment retry attempts.
		return "Stripe未能找到合适的支付方式，支付已取消"
	}

	return "未知"
}

func (s StripeSub) RequiresAction() bool {
	return s.Status == stripe.SubscriptionStatusIncomplete
}

type StripeInvoice struct {
	*stripe.Invoice
}

func (i StripeInvoice) CreationTime() chrono.Time {
	return chrono.TimeFrom(canonicalizeUnix(i.Created))
}

func (i StripeInvoice) BuildFtcPlan() (Plan, error) {
	if i.Lines == nil {
		return Plan{}, errors.New("empty lines")
	}

	if len(i.Lines.Data) == 0 {
		return Plan{}, errors.New("empty lines.data")
	}

	stripePlan := i.Lines.Data[0].Plan

	ftcPlan, err := GetFtcPlansWithStripe(i.Livemode).FindPlan(stripePlan.ID)
	if err != nil {
		return Plan{}, err
	}

	return ftcPlan.WithStripe(*stripePlan), nil
}

func (i StripeInvoice) Price() string {
	return fmt.Sprintf("%s%.2f", strings.ToUpper(string(i.Currency)), float64(i.Total/100))
}

// ReadableStatus turns stripe invoice status into human readable
// text.
// See https://stripe.com/docs/billing/invoices/workflow#workflow-overview
func (i StripeInvoice) ReadableStatus() string {
	switch i.Status {
	case stripe.InvoiceBillingStatusDraft:
		return "草稿"

	case stripe.InvoiceBillingStatusOpen:
		return "等待支付"

	case stripe.InvoiceBillingStatusPaid:
		return "已支付"

	case stripe.InvoiceBillingStatusUncollectible:
		return "无法收款"

	case stripe.InvoiceBillingStatusVoid:
		return "撤销"
	}

	return "未知"
}
