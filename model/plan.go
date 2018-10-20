package model

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// MemberTier represents membership tiers
type MemberTier string

// BillingCycle is an enum of billing cycles.
type BillingCycle string

// PaymentMethod lists supported payment channels.
type PaymentMethod string

const (
	// TierInvalid is a placeholder
	TierInvalid MemberTier = ""
	// TierStandard is the standard tier
	TierStandard MemberTier = "standard"
	// TierPremium is the premium tier
	TierPremium MemberTier = "premium"
	// CycleInvalid is a placeholder
	CycleInvalid BillingCycle = ""
	// Yearly bills every year
	Yearly BillingCycle = "year"
	// Monthly bills every month
	Monthly BillingCycle = "month"
	// Alipay supports taobao payment
	Alipay PaymentMethod = "alipay"
	// Wxpay supports wechat payment
	Wxpay PaymentMethod = "tenpay"
	// Stripe supports pay by stripe
	Stripe PaymentMethod = "stripe"
)

// NewTier returns a MemberTier.
// `key` either `standard` or `premium`
func NewTier(key string) (MemberTier, error) {
	switch key {
	case "standard":
		return TierStandard, nil

	case "premium":
		return TierPremium, nil

	default:
		return MemberTier(""), errors.New("Only standard and premium tier allowed")
	}
}

// NewCycle returns a new BillingCycle.
// `key` is either `year` or `month`.
func NewCycle(key string) (BillingCycle, error) {
	switch key {
	case "year":
		return Yearly, nil
	case "month":
		return Monthly, nil
	default:
		return CycleInvalid, errors.New("cycle must either be year or month")
	}
}

// Plan represents a subscription plan
type Plan struct {
	Tier        MemberTier
	Cycle       BillingCycle
	Price       float32
	ID          int // 10 for standard and 100 for premium
	Description string
}

// GetPriceCent calculates price in cent to be used for Wechat pay.
func (p Plan) GetPriceCent() int64 {
	return int64(p.Price * 100)
}

var plans = map[string]Plan{
	"standard_year": Plan{
		Tier:        TierStandard,
		Cycle:       Yearly,
		Price:       198.00,
		ID:          10,
		Description: "FT中文网 - 标准会员",
	},
	"standard_month": Plan{
		Tier:        TierStandard,
		Cycle:       Monthly,
		Price:       28.00,
		ID:          5,
		Description: "FT中文网 - 标准会员",
	},
	"premium_year": Plan{
		Tier:        TierPremium,
		Cycle:       Yearly,
		Price:       1998.00,
		ID:          100,
		Description: "FT中文网 - 高端会员",
	},
	"premium_month": Plan{
		Tier:        TierPremium,
		Cycle:       Monthly,
		Price:       280.00,
		ID:          50,
		Description: "FT中文网 - 高端会员",
	},
}

// CreateOrderID creates the order number based on the plan selected.
func CreateOrderID(p Plan) string {
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between [100, 999)
	rn := 100 + rand.Intn(999-100+1)

	return fmt.Sprintf("FT%03d%d%d", p.ID, rn, time.Now().Unix())
}

// NewPlan creates a new Plan instance depending on the member tier and billing cycle chosen.
// Returns error if member tier or billing cycyle are not the predefined ones.
func NewPlan(tier MemberTier, cycle BillingCycle) (Plan, error) {
	key := string(tier) + "_" + string(cycle)

	p, ok := plans[key]

	if !ok {
		return p, errors.New("subscription plan not found")
	}

	return p, nil
}
