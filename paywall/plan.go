package paywall

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
)

var (
	standardYearlyPlan = Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleYear,
		ListPrice:   258.00,
		NetPrice:    258.00,
		Description: "FT中文网 - 年度标准会员",
		CycleCount:  1,
		Currency:    "￥",
		ExtraDays:   1,
	}
	standardMonthlyPlan = Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleMonth,
		ListPrice:   28.00,
		NetPrice:    28.00,
		Description: "FT中文网 - 月度标准会员",
		CycleCount:  1,
		Currency:    "￥",
		ExtraDays:   1,
	}
	premiumYearlyPlan = Plan{
		Tier:        enum.TierPremium,
		Cycle:       enum.CycleYear,
		ListPrice:   1998.00,
		NetPrice:    1998.00,
		Description: "FT中文网 - 高端会员",
		CycleCount:  1,
		Currency:    "￥",
		ExtraDays:   1,
	}
)
var defaultPlans = FtcPlans{
	"standard_year":  standardYearlyPlan,
	"standard_month": standardMonthlyPlan,
	"premium_year":   premiumYearlyPlan,
}

var sandboxPlans = FtcPlans{
	"standard_year":  standardYearlyPlan.WithSandbox(),
	"standard_month": standardMonthlyPlan.WithSandbox(),
	"premium_year":   premiumYearlyPlan.WithSandbox(),
}

var stripeToFtcPlansTest = FtcPlans{
	"plan_FOdfeaqzczp6Ag": standardYearlyPlan.WithSandbox(),
	"plan_FOdgPTznDwHU4i": standardMonthlyPlan.WithSandbox(),
	"plan_FOde0uAr0V4WmT": premiumYearlyPlan.WithSandbox(),
}

var stripeToFtcPlansLive = FtcPlans{}

var stripeTestPlanIDs = StripePlans{
	"standard_year":  "plan_FOdfeaqzczp6Ag",
	"standard_month": "plan_FOdgPTznDwHU4i",
	"premium_year":   "plan_FOde0uAr0V4WmT",
}

var stripeLivePlanIDs = StripePlans{}

// Plan is a pricing plan.
// The list price is the price that buyers pay for your product or service without any discounts.
// The net price of a product or service is the actual price that customers pay for the product or service.
type Plan struct {
	Tier        enum.Tier  `json:"tier"` // This is product.
	Cycle       enum.Cycle `json:"cycle"`
	ListPrice   float64    `json:"listPrice"`
	NetPrice    float64    `json:"netPrice"`
	Description string     `json:"description"`
	CycleCount  int64      `json:"cycleCount"`
	Currency    string     `json:"currency"`
	ExtraDays   int64      `json:"extraDays"`
}

// WithSandbox returns the sandbox version of a plan.
func (p Plan) WithSandbox() Plan {
	p.NetPrice = 0.01
	return p
}

func (p Plan) BuildUpgradePlan(balance float64) Plan {
	if p.ListPrice >= balance {
		p.NetPrice = p.ListPrice - balance
	} else {
		p.CycleCount, p.ExtraDays = convertBalance(balance, p.ListPrice)
		p.NetPrice = 0
	}

	return p
}

func (p Plan) ProductID() string {
	return p.Tier.String() + "_" + p.Cycle.String()
}

func (p Plan) StripePrice() int64 {
	return int64(p.NetPrice * 100)
}

// FtcPlans maps a key to a FTC plan.
type FtcPlans map[string]Plan

// FindPlan picks a pricing plan from a group a pre-defined plans.
func (plans FtcPlans) FindPlan(tier, cycle string) (Plan, error) {
	key := tier + "_" + cycle

	p, ok := plans[key]

	if !ok {
		return p, fmt.Errorf("pricing plan for %s not found", key)
	}

	return p, nil
}

func (plans FtcPlans) GetPlanByID(id string) (Plan, error) {
	p, ok := plans[id]
	if !ok {
		return p, fmt.Errorf("pricing plan for %s not found", id)
	}

	return p, nil
}

type StripePlans map[string]string

func (plans StripePlans) FindPlanID(key string) (string, error) {
	id, ok := plans[key]
	if !ok {
		return id, fmt.Errorf("stripe plan id for %s not found", key)
	}

	return id, nil
}

func GetFtcPlans(sandbox bool) FtcPlans {
	if sandbox {
		return sandboxPlans
	}

	return defaultPlans
}

func GetStripeToFtcPlans(live bool) FtcPlans {
	if live {
		return stripeToFtcPlansLive
	}

	return stripeToFtcPlansTest
}

func GetStripePlans(live bool) StripePlans {
	if live {
		return stripeLivePlanIDs
	}

	return stripeTestPlanIDs
}
