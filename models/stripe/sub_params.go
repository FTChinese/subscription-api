package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
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
func (sp SubParams) NewMembership(id reader.MemberID, ss *stripe.Subscription) subscription.Membership {

	periodEnd := CanonicalizeUnix(ss.CurrentPeriodEnd)
	status, _ := subscription.ParseSubStatus(string(ss.Status))

	return subscription.Membership{
		ID:       null.StringFrom(subscription.GenerateMembershipIndex()),
		MemberID: id,
		BasePlan: plan.BasePlan{
			Tier:  sp.Tier,
			Cycle: sp.Cycle,
		},
		ExpireDate:    chrono.DateFrom(periodEnd.AddDate(0, 0, 1)),
		PaymentMethod: enum.PayMethodStripe,
		StripeSubID:   null.StringFrom(ss.ID),
		StripePlanID:  null.StringFrom(sp.GetStripePlanID()),
		AutoRenew:     !ss.CancelAtPeriodEnd,
		Status:        status,
	}
}

// UpdateMembership updates an existing membership for a new stripe subscription.
func (sp SubParams) UpdateMembership(m subscription.Membership, ss *stripe.Subscription) subscription.Membership {
	periodEnd := CanonicalizeUnix(ss.CurrentPeriodEnd)
	status, _ := subscription.ParseSubStatus(string(ss.Status))

	m.GenerateID()
	m.Tier = sp.Tier
	m.Cycle = sp.Cycle
	m.ExpireDate = chrono.DateFrom(periodEnd.AddDate(0, 0, 1))
	m.PaymentMethod = enum.PayMethodStripe
	m.StripeSubID = null.StringFrom(ss.ID)
	m.StripePlanID = null.StringFrom(sp.GetStripePlanID())
	m.AutoRenew = !ss.CancelAtPeriodEnd
	m.Status = status

	return m
}

// RefreshMembership refreshes an existing valid stripe membership.
func RefreshMembership(m subscription.Membership, ss *stripe.Subscription) subscription.Membership {
	m.GenerateID()
	periodEnd := CanonicalizeUnix(ss.CurrentPeriodEnd)

	m.ExpireDate = chrono.DateFrom(periodEnd.AddDate(0, 0, 1))
	m.AutoRenew = !ss.CancelAtPeriodEnd
	m.Status, _ = subscription.ParseSubStatus(string(ss.Status))

	return m
}
