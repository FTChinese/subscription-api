package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/models/plan"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
)

// SubParams defines the payload when requesting to create/upgrade a stripe subscription.
type SubParams struct {
	plan.BasePlan
	Customer             string      `json:"customer"`
	Coupon               null.String `json:"coupon"`
	DefaultPaymentMethod null.String `json:"defaultPaymentMethod"`
	IdempotencyKey       string      `json:"idempotency"`
	live                 bool
	plan                 plan.Plan
}

// SetLive determines which stripe plan id will be used.
func (sp *SubParams) SetLive(live bool) *SubParams {
	sp.live = live
	return sp
}

// GetFtcPlan tries to find out which Plan user chose.
func (sp *SubParams) GetFtcPlan() (plan.Plan, error) {
	if sp.plan.Tier != enum.TierNull && sp.plan.Cycle != enum.CycleNull {
		return sp.plan, nil
	}

	ftcPlan, err := plan.FindPlan(sp.Tier, sp.Cycle)
	if err != nil {
		return plan.Plan{}, nil
	}

	sp.plan = ftcPlan

	return ftcPlan, nil
}

// GetStripePlanID returns the stripe plan id under current environment.
func (sp SubParams) GetStripePlanID() string {
	return sp.plan.GetStripePlanID(sp.live)
}

// NewMembership creates a new membership for stripe subscription
func (sp SubParams) NewMembership(id reader.MemberID, ss *stripe.Subscription) subs.Membership {

	periodEnd := CanonicalizeUnix(ss.CurrentPeriodEnd)
	status, _ := enum.ParseSubsStatus(string(ss.Status))

	return subs.Membership{
		MemberID: id,
		Edition: product.Edition{
			Tier:  sp.Tier,
			Cycle: sp.Cycle,
		},
		ExpireDate:    chrono.DateFrom(periodEnd.AddDate(0, 0, 1)),
		PaymentMethod: enum.PayMethodStripe,
		StripeSubsID:  null.StringFrom(ss.ID),
		StripePlanID:  null.StringFrom(sp.GetStripePlanID()),
		AutoRenewal:   !ss.CancelAtPeriodEnd,
		Status:        status,
	}
}

// UpdateMembership updates an existing membership for a new stripe subscription.
func (sp SubParams) UpdateMembership(m subs.Membership, ss *stripe.Subscription) subs.Membership {
	periodEnd := CanonicalizeUnix(ss.CurrentPeriodEnd)
	status, _ := enum.ParseSubsStatus(string(ss.Status))

	m.Tier = sp.Tier
	m.Cycle = sp.Cycle
	m.ExpireDate = chrono.DateFrom(periodEnd.AddDate(0, 0, 1))
	m.PaymentMethod = enum.PayMethodStripe
	m.StripeSubsID = null.StringFrom(ss.ID)
	m.StripePlanID = null.StringFrom(sp.GetStripePlanID())
	m.AutoRenewal = !ss.CancelAtPeriodEnd
	m.Status = status

	return m
}

// RefreshMembership refreshes an existing valid stripe membership.
func RefreshMembership(m subs.Membership, ss *stripe.Subscription) subs.Membership {
	periodEnd := CanonicalizeUnix(ss.CurrentPeriodEnd)

	m.ExpireDate = chrono.DateFrom(periodEnd.AddDate(0, 0, 1))
	m.AutoRenewal = !ss.CancelAtPeriodEnd
	m.Status, _ = enum.ParseSubsStatus(string(ss.Status))

	return m
}
