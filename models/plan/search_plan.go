package plan

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

var (
	standardMonthlyPlan = Plan{
		BasePlan: BasePlan{
			Tier:       enum.TierStandard,
			Cycle:      enum.CycleMonth,
			LegacyTier: null.IntFrom(5),
		},
		ListPrice: 28.00,
		NetPrice:  28.00,
		Price:     28.00,
		Amount:    28.00,
		Duration: Duration{
			CycleCount: 1,
			ExtraDays:  1,
		},
		Title:            "FT中文网 - 月度标准会员",
		Currency:         "cny",
		stripeLivePlanID: "plan_FXZYLOEbcvj5Tx",
		stripeTestPlanID: "plan_FOdgPTznDwHU4i",
		AppleProductID:   "com.ft.ftchinese.mobile.subscription.member.monthly",
	}

	standardYearlyPlan = Plan{
		BasePlan: BasePlan{
			Tier:       enum.TierStandard,
			Cycle:      enum.CycleYear,
			LegacyTier: null.IntFrom(10),
		},
		ListPrice: 258.00,
		NetPrice:  258.00,
		Price:     258.00,
		Amount:    258.00,
		Duration: Duration{
			CycleCount: 1,
			ExtraDays:  1,
		},
		Title:            "FT中文网 - 年度标准会员",
		Currency:         "cny",
		stripeLivePlanID: "plan_FXZZUEDpToPlZK",
		stripeTestPlanID: "plan_FOdfeaqzczp6Ag",
		AppleProductID:   "com.ft.ftchinese.mobile.subscription.member",
	}

	premiumYearlyPlan = Plan{
		BasePlan: BasePlan{
			Tier:       enum.TierPremium,
			Cycle:      enum.CycleYear,
			LegacyTier: null.IntFrom(100),
		},
		ListPrice: 1998.00,
		NetPrice:  1998.00,
		Price:     1998.00,
		Amount:    1998.00,
		Duration: Duration{
			CycleCount: 1,
			ExtraDays:  1,
		},
		Title:            "FT中文网 - 年度高端会员",
		Currency:         "cny",
		stripeLivePlanID: "plan_FXZbv1cDTsUKOg",
		stripeTestPlanID: "plan_FOde0uAr0V4WmT",
		AppleProductID:   "com.ft.ftchinese.mobile.subscription.vip",
	}
)

// FtcPlans maps a key to a FTC plan.
type FtcPlans map[string]Plan

// FindPlan searches a plan by a key.
func (plans FtcPlans) FindPlan(id string) (Plan, error) {
	p, ok := plans[id]
	if !ok {
		return p, fmt.Errorf("pricing plan for %s not found", id)
	}

	return p, nil
}

// Index FTC plans by plan name.
var ftcPlansLive = FtcPlans{
	"standard_year":  standardYearlyPlan,
	"standard_month": standardMonthlyPlan,
	"premium_year":   premiumYearlyPlan,
}

func GetFtcPlans(live bool) FtcPlans {
	if live {
		return ftcPlansLive
	}

	plans := FtcPlans{}

	for key, plan := range ftcPlansLive {
		p := plan.withSandboxPrice()
		plans[key] = p
	}

	return plans
}

func FindFtcPlan(id string) (Plan, error) {
	return ftcPlansLive.FindPlan(id)
}

func FindPlan(tier enum.Tier, cycle enum.Cycle) (Plan, error) {
	key := tier.String() + "_" + cycle.String()

	return ftcPlansLive.FindPlan(key)
}

var stripeLivePlans = FtcPlans{
	standardMonthlyPlan.stripeLivePlanID: standardMonthlyPlan,
	standardYearlyPlan.stripeLivePlanID:  standardYearlyPlan,
	premiumYearlyPlan.stripeLivePlanID:   premiumYearlyPlan,
}

var stripeTestPlans = FtcPlans{
	standardMonthlyPlan.stripeTestPlanID: standardMonthlyPlan,
	standardYearlyPlan.stripeTestPlanID:  standardYearlyPlan,
	premiumYearlyPlan.stripeTestPlanID:   premiumYearlyPlan,
}

// FindPlanForStripe finds a Plan for a stripe ID.
// This will replace the above complex and confusing ways
// of mapping stripe id to our Plan.
func FindPlanForStripe(id string, live bool) (Plan, error) {
	if live {
		return stripeLivePlans.FindPlan(id)
	}

	return stripeTestPlans.FindPlan(id)
}

var appleProductPlans = FtcPlans{
	standardMonthlyPlan.AppleProductID: standardMonthlyPlan,
	standardYearlyPlan.AppleProductID:  standardYearlyPlan,
	premiumYearlyPlan.AppleProductID:   premiumYearlyPlan,
}

func FindPlanForApple(id string) (Plan, error) {
	return appleProductPlans.FindPlan(id)
}

func AppleProductExists(id string) bool {
	_, ok := appleProductPlans[id]

	return ok
}
