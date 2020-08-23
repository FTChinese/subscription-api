package stripe

import (
	"database/sql"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/sub"
)

// SubsInput is the request body to create a new subscription.
type SubsInput struct {
	product.Edition
	FtcID                string      `json:"-"` // Acquired from header.
	PlanID               string      `json:"-"`
	CustomerID           string      `json:"customer"`
	CouponID             null.String `json:"coupon"`
	DefaultPaymentMethod null.String `json:"defaultPaymentMethod"`
	IdempotencyKey       string      `json:"idempotency"`
}

func NewSubsInput(ftcID string) SubsInput {
	return SubsInput{
		FtcID: ftcID,
	}
}

func (i SubsInput) WithPlanID(live bool) (SubsInput, error) {
	p, err := stripePlans.findByEdition(i.NamedKey(), live)
	if err != nil {
		return i, sql.ErrNoRows
	}

	i.PlanID = p.PlanID

	return i, nil
}

// Create calls stripe API to create a new subscription.
func (i SubsInput) CreateSubs() (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(i.CustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: stripe.String(i.PlanID),
			},
		},
	}

	// {
	// "status":400,
	// "message":"Idempotent key length is 0 characters long, which is outside accepted lengths. Idempotent Keys must be 1-255 characters long. If you're looking for a decent generator, try using a UUID defined by IETF RFC 4122.",
	// "request_id":"req_O6zILK5QEVpViw",
	// "type":"idempotency_error"
	// }
	if i.IdempotencyKey != "" {
		params.SetIdempotencyKey(i.IdempotencyKey)
	}

	if i.CouponID.Valid {
		params.Coupon = stripe.String(i.DefaultPaymentMethod.String)
	}

	if i.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(i.DefaultPaymentMethod.String)
	}

	// Call API.
	return sub.New(params)
}

func (i SubsInput) UpgradeSubs(subsID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: stripe.String(i.PlanID),
			},
		},
	}

	params.AddExpand("latest_invoice.payment_intent")

	if i.IdempotencyKey != "" {
		params.IdempotencyKey = stripe.String(i.IdempotencyKey)
	}

	if i.CouponID.Valid {
		params.Coupon = stripe.String(i.CouponID.String)
	}

	if i.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(i.DefaultPaymentMethod.String)
	}

	params.SetIdempotencyKey(i.IdempotencyKey)
	return sub.Update(subsID, params)
}

// NewMembership creates a new membership for stripe subscription
func (i SubsInput) NewMembership(ss *stripe.Subscription) reader.Membership {

	periodEnd := CanonicalizeUnix(ss.CurrentPeriodEnd)
	status, _ := enum.ParseSubsStatus(string(ss.Status))

	return reader.Membership{
		MemberID: reader.MemberID{
			CompoundID: "",
			FtcID:      null.StringFrom(i.FtcID),
			UnionID:    null.String{},
		}.MustNormalize(),
		Edition:       i.Edition,
		ExpireDate:    chrono.DateFrom(periodEnd.AddDate(0, 0, 1)),
		PaymentMethod: enum.PayMethodStripe,
		StripeSubsID:  null.StringFrom(ss.ID),
		StripePlanID:  null.StringFrom(i.PlanID),
		AutoRenewal:   !ss.CancelAtPeriodEnd,
		Status:        status,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
	}
}

// UpdateMembership updates an existing membership for a new stripe subscription.
func (i SubsInput) UpdateMembership(m reader.Membership, ss *stripe.Subscription) reader.Membership {
	periodEnd := CanonicalizeUnix(ss.CurrentPeriodEnd)
	status, _ := enum.ParseSubsStatus(string(ss.Status))

	m.Tier = i.Tier
	m.Cycle = i.Cycle
	m.ExpireDate = chrono.DateFrom(periodEnd.AddDate(0, 0, 1))
	m.PaymentMethod = enum.PayMethodStripe
	m.FtcPlanID = null.String{}
	m.StripeSubsID = null.StringFrom(ss.ID)
	m.StripePlanID = null.StringFrom(i.PlanID)
	m.AutoRenewal = !ss.CancelAtPeriodEnd
	m.Status = status
	m.AppleSubsID = null.String{}
	m.B2BLicenceID = null.String{}

	return m
}

func GetSubscription(subsID string) (*stripe.Subscription, error) {
	return sub.Get(subsID, nil)
}

// RefreshMembership refreshes an existing valid stripe membership.
func RefreshMembership(m reader.Membership, ss *stripe.Subscription) reader.Membership {
	periodEnd := CanonicalizeUnix(ss.CurrentPeriodEnd)

	m.ExpireDate = chrono.DateFrom(periodEnd.AddDate(0, 0, 1))
	m.AutoRenewal = !ss.CancelAtPeriodEnd
	m.Status, _ = enum.ParseSubsStatus(string(ss.Status))

	return m
}
