package util

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
	// Standard is the standard tier
	Standard MemberTier = "standard"
	// Premium is the premium tier
	Premium MemberTier = "premium"
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

var tiers = map[string]MemberTier{
	"standard": Standard,
	"premium":  Premium,
}

// NewTier tests if a tier eixsts and returns it.
// `key` is the request url's `tier` part
func NewTier(key string) (MemberTier, bool) {
	t, ok := tiers[key]

	return t, ok
}

var cycles = map[string]BillingCycle{
	"year":  Yearly,
	"month": Monthly,
}

// NewCycle returns a BillingCycle if key is one of year or month.
// `key` is request url's `cycle` part.
func NewCycle(key string) (BillingCycle, bool) {
	c, ok := cycles[key]

	return c, ok
}

// Plan represents a subscription plan
type Plan struct {
	Tier        MemberTier
	Cycle       BillingCycle
	Price       int64
	ID          int // 10 for standard and 100 for premium
	Description string
}

// GetPriceCent calculates price in cent to be used for Wechat pay.
func (p Plan) GetPriceCent() int64 {
	return p.Price * 100
}

var plans = map[string]Plan{
	"standard_year": Plan{
		Tier:        Standard,
		Cycle:       Yearly,
		Price:       198,
		ID:          10,
		Description: "FT中文网 - 标准会员",
	},
	"standard_month": Plan{
		Tier:        Standard,
		Cycle:       Monthly,
		Price:       28,
		ID:          5,
		Description: "FT中文网 - 标准会员",
	},
	"premium_year": Plan{
		Tier:        Premium,
		Cycle:       Yearly,
		Price:       1998,
		ID:          100,
		Description: "FT中文网 - 高端会员",
	},
	"premium_month": Plan{
		Tier:        Premium,
		Cycle:       Monthly,
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
