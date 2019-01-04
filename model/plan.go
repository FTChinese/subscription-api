package model

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"gitlab.com/ftchinese/subscription-api/enum"
)

// Plan represents a subscription plan
type Plan struct {
	Tier        enum.Tier  `json:"tier"`
	Cycle       enum.Cycle `json:"cycle"`
	Price       float64    `json:"price"`
	ID          int        `json:"id"` // 10 for standard and 100 for premium
	Description string     `json:"description"`
	Ignore      bool       `json:"ignore,omitempty"`
}

// GetPriceCent calculates price in cent to be used for Wechat pay.
func (p Plan) GetPriceCent() int64 {
	return int64(p.Price * 100)
}

// GetPriceString formats price for alipay
func (p Plan) GetPriceString() string {
	return strconv.FormatFloat(p.Price, 'f', 2, 32)
}

// OrderID generates an FT order id based
// on the plan's id, a random number between 100 to 999,
// and unix timestamp.
func (p Plan) OrderID() string {
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between [100, 999)
	rn := 100 + rand.Intn(999-100+1)

	return fmt.Sprintf("FT%03d%d%d", p.ID, rn, time.Now().Unix())
}

// Pricing defines a collection pricing plan.
type Pricing map[string]Plan

// FindPlan picks a pricing plan from a group a pre-defined plans.
func (plans Pricing) FindPlan(tier, cycle string) (Plan, error) {
	key := tier + "_" + cycle

	p, ok := plans[key]

	if !ok {
		return p, errors.New("subscription plan not found")
	}

	return p, nil
}

// DefaultPlans is the default subscription. No discount.
var DefaultPlans = Pricing{
	"standard_year": Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleYear,
		Price:       258.00,
		ID:          10,
		Description: "FT中文网 - 年度标准会员",
	},
	"standard_month": Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleMonth,
		Price:       28.00,
		ID:          5,
		Description: "FT中文网 - 月度标准会员",
	},
	"premium_year": Plan{
		Tier:        enum.TierPremium,
		Cycle:       enum.CycleYear,
		Price:       1998.00,
		ID:          100,
		Description: "FT中文网 - 高端会员",
	},
}

// SandboxPlans is used by sandbox for testing.
var SandboxPlans = Pricing{
	"standard_year": Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleYear,
		Price:       0.01,
		ID:          10,
		Description: "FT中文网 - 年度标准会员",
	},
	"standard_month": Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleMonth,
		Price:       0.01,
		ID:          5,
		Description: "FT中文网 - 月度标准会员",
	},
	"premium_year": Plan{
		Tier:        enum.TierPremium,
		Cycle:       enum.CycleYear,
		Price:       0.01,
		ID:          100,
		Description: "FT中文网 - 高端会员",
	},
}

// LoadCurrentPlans get default plans or promo plans depending on current time.
func (env Env) LoadCurrentPlans() Pricing {

	// First, check if cache has any promotion schedules
	promo, found := env.PromoFromCache()

	// If no cache is found, use default ones.
	if !found || !promo.isInEffect() {
		logger.WithField("location", "GetCurrentPlans").Info("Use defualt plans")
		return DefaultPlans
	}

	return promo.Plans
}

// FindPlan picks a Plan instance depending
// on the member tier and billing cycle.
// tier is an enum: standard | premium.
// cycle is an enum: year | month
// Returns error if member tier or billing cycyle are not in the predefined ones.
func (env Env) FindPlan(tier, cycle string) (Plan, error) {
	key := tier + "_" + cycle

	plans := env.LoadCurrentPlans()
	p, ok := plans[key]

	if !ok {
		return p, errors.New("subscription plan not found")
	}

	return p, nil
}
