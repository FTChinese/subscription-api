package paywall

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
	"time"
)

type StripeSubParams struct {
	Coordinate
	Customer             string      `json:"customer"`
	Coupon               null.String `json:"coupon"`
	DefaultPaymentMethod null.String `json:"defaultPaymentMethod"`
	IdempotencyKey       string      `json:"idempotency"`
	planID               string
}

func (p *StripeSubParams) SetStripePlanID(live bool) error {
	plan, err := GetFtcPlans(live).FindPlan(p.PlanID())
	if err != nil {
		return nil
	}

	p.planID = plan.StripeID

	return nil
}

func (p StripeSubParams) GetPlanID() string {
	return p.planID
}

func (p StripeSubParams) GetStripePlanID(live bool) (string, error) {
	plan, err := GetFtcPlans(live).FindPlan(p.PlanID())
	if err != nil {
		return "", nil
	}

	return plan.StripeID, nil
}

type StripeSub struct {
	CancelAtPeriodEnd  bool          `json:"cancelAtPeriodEnd"`
	Created            chrono.Time   `json:"created"`
	CurrentPeriodEnd   chrono.Time   `json:"currentPeriodEnd"`
	CurrentPeriodStart chrono.Time   `json:"currentPeriodStart"`
	CustomerID         string        `json:"customerId"`
	EndedAt            chrono.Time   `json:"endedAt"`
	ID                 string        `json:"id"`
	LatestInvoice      StripeInvoice `json:"latestInvoice"`
	Livemode           bool          `json:livemode`
	Plan               stripe.Plan   `json:"plan"`
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
	plan, _ := extractStripePlan(s)

	return StripeSub{
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		Created:            chrono.TimeFrom(canonicalizeUnix(s.Created)),
		CurrentPeriodEnd:   chrono.TimeFrom(canonicalizeUnix(s.CurrentPeriodEnd)),
		CurrentPeriodStart: chrono.TimeFrom(canonicalizeUnix(s.CurrentPeriodStart)),
		CustomerID:         s.Customer.ID,
		EndedAt:            chrono.TimeFrom(canonicalizeUnix(s.EndedAt)),
		ID:                 s.ID,
		LatestInvoice:      NewStripeInvoice(s.LatestInvoice),
		Livemode:           s.Livemode,
		Plan:               plan,
		StartDate:          chrono.TimeFrom(canonicalizeUnix(s.StartDate)),
		Status:             s.Status,
	}
}

func (s StripeSub) BuildFtcPlan() (Plan, error) {
	ftcPlan, err := GetStripeToFtcPlans(s.Livemode).FindPlan(s.Plan.ID)
	if err != nil {
		return ftcPlan, err
	}

	return ftcPlan.WithStripe(s.Plan), nil
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

func extractStripePlan(s *stripe.Subscription) (stripe.Plan, error) {
	if s.Plan != nil {
		return *s.Plan, nil
	}

	if s.Items == nil {
		return stripe.Plan{}, errors.New("stripe subscription items are empty")
	}
	if len(s.Items.Data) == 0 {
		return stripe.Plan{}, errors.New("stripe subscription items are empty")
	}

	if s.Items.Data[0].Plan == nil {
		return stripe.Plan{}, errors.New("stripe subscription plan is nil")
	}

	return *s.Items.Data[0].Plan, nil
}
