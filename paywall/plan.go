package paywall

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/stripe/stripe-go"
	"strings"
)

var (
	standardMonthlyPlan = Plan{
		Coordinate: Coordinate{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleMonth,
		},
		ListPrice:  28.00,
		NetPrice:   28.00,
		Title:      "FT中文网 - 月度标准会员",
		CycleCount: 1,
		Currency:   "cny",
		ExtraDays:  1,
		StripeID:   "",
	}

	standardYearlyPlan = Plan{
		Coordinate: Coordinate{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleYear,
		},
		ListPrice:  258.00,
		NetPrice:   258.00,
		Title:      "FT中文网 - 年度标准会员",
		CycleCount: 1,
		Currency:   "cny",
		ExtraDays:  1,
		StripeID:   "",
	}

	premiumYearlyPlan = Plan{
		Coordinate: Coordinate{
			Tier:  enum.TierPremium,
			Cycle: enum.CycleYear,
		},
		ListPrice:  1998.00,
		NetPrice:   1998.00,
		Title:      "FT中文网 - 高端会员",
		CycleCount: 1,
		Currency:   "cny",
		ExtraDays:  1,
		StripeID:   "",
	}
)

var stripeTestPlanIDs = map[string]string{
	"standard_year":  "plan_FOdfeaqzczp6Ag",
	"standard_month": "plan_FOdgPTznDwHU4i",
	"premium_year":   "plan_FOde0uAr0V4WmT",
}

var ordinals = map[string]int{
	"standard_month": 5,
	"standard_year":  10,
	"premium_year":   100,
}

var ftcPlansLive = FtcPlans{
	"standard_year":  standardYearlyPlan,
	"standard_month": standardMonthlyPlan,
	"premium_year":   premiumYearlyPlan,
}

var ftcPlansSandbox = buildSandboxPlans()

func buildSandboxPlans() FtcPlans {
	plans := FtcPlans{}

	for key, plan := range ftcPlansLive {
		p := plan.withSandbox()
		p.StripeID = stripeTestPlanIDs[key]
		plans[key] = p
	}

	return plans
}

// Coordinate includes a product and a plan billing cycle to identify the plan to subscribe.
type Coordinate struct {
	Tier  enum.Tier  `json:"tier"`
	Cycle enum.Cycle `json:"cycle"`
}

// PlanID create a unique name for the point in the plane.
func (c Coordinate) PlanID() string {
	return c.Tier.String() + "_" + c.Cycle.String()
}

func (c Coordinate) GetOrdinal() int {
	return ordinals[c.PlanID()]
}

// Plan is a pricing plan.
// The list price is the price that buyers pay for your product or service without any discounts.
// The net price of a product or service is the actual price that customers pay for the product or service.
type Plan struct {
	Coordinate
	ListPrice  float64 `json:"listPrice"`
	NetPrice   float64 `json:"netPrice"`
	Title      string  `json:"description"`
	CycleCount int64   `json:"cycleCount"`
	Currency   string  `json:"currency"`
	ExtraDays  int64   `json:"extraDays"`
	StripeID   string  `json:"-"`
}

// withSandbox returns the sandbox version of a plan.
func (p Plan) withSandbox() Plan {
	p.NetPrice = 0.01
	return p
}

func (p Plan) WithUpgrade(balance float64) Plan {
	if p.ListPrice >= balance {
		p.NetPrice = p.ListPrice - balance
	} else {
		p.CycleCount, p.ExtraDays = convertBalance(balance, p.ListPrice)
		p.NetPrice = 0
	}

	return p
}

func (p Plan) WithStripe(sp stripe.Plan) Plan {
	p.ListPrice = float64(sp.Amount / 100)
	p.NetPrice = p.ListPrice
	p.Currency = string(sp.Currency)

	return p
}

// Desc is used for displaying to user.
// The price show here is not the final price user paid.
func (p Plan) Desc() string {
	return fmt.Sprintf(
		"%s %s%.2f/%s",
		p.Tier.StringCN(),
		strings.ToUpper(p.Currency),
		p.ListPrice,
		p.Cycle.StringCN(),
	)
}

// FtcPlans maps a key to a FTC plan.
type FtcPlans map[string]Plan

func (plans FtcPlans) FindPlan(id string) (Plan, error) {
	p, ok := plans[id]
	if !ok {
		return p, fmt.Errorf("pricing plan for %s not found", id)
	}

	return p, nil
}

var stripeToFtcPlansTest = FtcPlans{
	"plan_FOdfeaqzczp6Ag": standardYearlyPlan.withSandbox(),
	"plan_FOdgPTznDwHU4i": standardMonthlyPlan.withSandbox(),
	"plan_FOde0uAr0V4WmT": premiumYearlyPlan.withSandbox(),
}

var stripeToFtcPlansLive = FtcPlans{}

func GetFtcPlans(live bool) FtcPlans {
	if live {
		return ftcPlansLive
	}

	return ftcPlansSandbox
}

func GetStripeToFtcPlans(live bool) FtcPlans {
	if live {
		return stripeToFtcPlansLive
	}

	return stripeToFtcPlansTest
}
