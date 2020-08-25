package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/stripe/stripe-go"
	"time"
)

// Subscription is a reduced version of stripe.Subscription.
// Used as response when client asks for subscription data.
type Subscription struct {
	AccountID reader.MemberID `json:"-"`
	product.Edition
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
// Unix 0 represent year 1970, while Golang's zero time is actually 0.
func CanonicalizeUnix(s int64) time.Time {
	if s > 0 {
		return time.Unix(s, 0)
	}

	return time.Time{}
}

func NewSubsResponse(s *stripe.Subscription) Subscription {
	if s == nil {
		return Subscription{}
	}

	return Subscription{
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		Created:            chrono.TimeFrom(CanonicalizeUnix(s.Created)),
		CurrentPeriodEnd:   chrono.TimeFrom(CanonicalizeUnix(s.CurrentPeriodEnd)),
		CurrentPeriodStart: chrono.TimeFrom(CanonicalizeUnix(s.CurrentPeriodStart)),
		CustomerID:         s.Customer.ID,
		EndedAt:            chrono.TimeFrom(CanonicalizeUnix(s.EndedAt)),
		ID:                 s.ID,
		LatestInvoiceID:    s.LatestInvoice.ID,
		Livemode:           s.Livemode,
		StartDate:          chrono.TimeFrom(CanonicalizeUnix(s.StartDate)),
		Status:             s.Status,
	}
}

func (s Subscription) ReadableStatus() string {
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

func (s Subscription) RequiresAction() bool {
	return s.Status == stripe.SubscriptionStatusIncomplete
}
