package paywall

import (
	"errors"

	"gitlab.com/ftchinese/next-api/enum"
)

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
var defaultPlans = Pricing{
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
var sandboxPlans = Pricing{
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

// GetDefaultPricing returns the default pricing plans.
func (env Env) GetDefaultPricing() Pricing {
	return defaultPlans
}

// GetCurrentPricing get current effective pricing plans.
func (env Env) GetCurrentPricing() Pricing {
	return defaultPlans
}

// GetSandboxPricing returns the pricing plans for sandbox.
func (env Env) GetSandboxPricing() Pricing {
	return sandboxPlans
}
