package paywall

import (
	"fmt"

	"github.com/FTChinese/go-rest/enum"
)

// Pricing defines a collection pricing plan.
type Pricing map[string]Plan

// FindPlan picks a pricing plan from a group a pre-defined plans.
func (plans Pricing) FindPlan(tier, cycle string) (Plan, error) {
	key := tier + "_" + cycle

	p, ok := plans[key]

	if !ok {
		return p, fmt.Errorf("pricing plan for %s not found", key)
	}

	return p, nil
}

// DefaultPlans is the default subscription. No discount.
var defaultPlans = Pricing{
	"standard_year": Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleYear,
		ListPrice:   258.00,
		NetPrice:    258.00,
		Description: "FT中文网 - 年度标准会员",
	},
	"standard_month": Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleMonth,
		ListPrice:   28.00,
		NetPrice:    28.00,
		Description: "FT中文网 - 月度标准会员",
	},
	"premium_year": Plan{
		Tier:        enum.TierPremium,
		Cycle:       enum.CycleYear,
		ListPrice:   1998.00,
		NetPrice:    1998.00,
		Description: "FT中文网 - 高端会员",
	},
}

// SandboxPlans is used by sandbox for testing.
var sandboxPlans = Pricing{
	"standard_year": Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleYear,
		ListPrice:   258.00,
		NetPrice:    0.01,
		Description: "FT中文网 - 年度标准会员",
	},
	"standard_month": Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleMonth,
		ListPrice:   28.00,
		NetPrice:    0.01,
		Description: "FT中文网 - 月度标准会员",
	},
	"premium_year": Plan{
		Tier:        enum.TierPremium,
		Cycle:       enum.CycleYear,
		ListPrice:   1998.00,
		NetPrice:    0.01,
		Description: "FT中文网 - 高端会员",
	},
}

// GetDefaultPricing returns the default pricing plans.
func GetDefaultPricing() Pricing {
	return defaultPlans
}

// GetSandboxPricing returns the pricing plans for sandbox.
func GetSandboxPricing() Pricing {
	return sandboxPlans
}
