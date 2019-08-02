package paywall

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/stripe/stripe-go"
	"math"
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
		StripeID:   "plan_FXZYLOEbcvj5Tx",
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
		StripeID:   "plan_FXZZUEDpToPlZK",
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
		StripeID:   "plan_FXZbv1cDTsUKOg",
	}
)

var ordinals = map[string]int{
	"standard_month": 5,
	"standard_year":  10,
	"premium_year":   100,
}

// Coordinate includes a product and a plan billing cycle to identify the plan to subscribe.
type Coordinate struct {
	Tier  enum.Tier  `json:"tier"`
	Cycle enum.Cycle `json:"cycle"`
}

// NamedKey create a unique name for the point in the plane.
func (c Coordinate) NamedKey() string {
	return c.Tier.String() + "_" + c.Cycle.String()
}

func (c Coordinate) GetOrdinal() int {
	return ordinals[c.NamedKey()]
}

// QuoRem returns the integer quotient and remainder of x/y
func QuoRem(x, y float64) (int64, float64) {
	var q int64

	for x > y {
		q = q + 1
		x = x - y
	}

	return q, x
}

type CycleQuantity struct {
	Count     int64
	ExtraDays int64
}

// Plan is a pricing plan.
// The list price is the price that buyers pay for your product or service without any discounts.
// The net price of a product or service is the actual price that customers pay for the product or service.
type Plan struct {
	Coordinate
	ListPrice  float64 `json:"listPrice"`
	NetPrice   float64 `json:"netPrice"`
	CycleCount int64   `json:"cycleCount"`
	Currency   string  `json:"currency"`
	ExtraDays  int64   `json:"extraDays"`
	Title      string  `json:"description"`
	StripeID   string  `json:"-"`
}

// withSandbox returns the sandbox version of a plan.
func (p Plan) withSandbox() Plan {
	p.NetPrice = 0.01
	return p
}

// WithUpgrade creates an upgrading plan.
func (p Plan) WithUpgrade(balance float64) Plan {

	if balance < p.NetPrice {
		p.NetPrice = p.NetPrice - balance
	} else {
		p.NetPrice = 0
	}

	q := p.CalculateConversion(balance)

	p.CycleCount = q.Count
	p.ExtraDays = q.ExtraDays

	return p
}

// ConvertBalance checks to see how many cycles and extra
// days a user's balance could be exchanged.
func (p Plan) CalculateConversion(balance float64) CycleQuantity {

	if balance <= p.NetPrice {
		return CycleQuantity{
			Count:     1,
			ExtraDays: 1,
		}
	}

	cycles, r := QuoRem(balance, p.NetPrice)

	days := math.Ceil(r * 365 / p.NetPrice)

	return CycleQuantity{
		Count:     cycles,
		ExtraDays: int64(days),
	}
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

// WithStripeIndex changes the map key to stripe plan id.
func (plans FtcPlans) WithStripeIndex() FtcPlans {
	ret := FtcPlans{}
	for _, v := range plans {
		ret[v.StripeID] = v
	}

	return ret
}

// FindPlan searches a plan by a key.
func (plans FtcPlans) FindPlan(id string) (Plan, error) {
	p, ok := plans[id]
	if !ok {
		return p, fmt.Errorf("pricing plan for %s not found", id)
	}

	return p, nil
}

// buildSandboxPlans build FTC plans for sandbox environemnt.
// It differs from live environment in terms of NetPrice and StripeID.
func buildSandboxPlans() FtcPlans {
	plans := FtcPlans{}

	for key, plan := range ftcPlansLive {
		p := plan.withSandbox()
		p.StripeID = stripePlanIDsTest[key]
		plans[key] = p
	}

	return plans
}

// Use a plan name to find a stripe plan's id.
var stripePlanIDsTest = map[string]string{
	"standard_month": "plan_FOdgPTznDwHU4i",
	"standard_year":  "plan_FOdfeaqzczp6Ag",
	"premium_year":   "plan_FOde0uAr0V4WmT",
}

// Index FTC plans by plan name.
var ftcPlansLive = FtcPlans{
	"standard_year":  standardYearlyPlan,
	"standard_month": standardMonthlyPlan,
	"premium_year":   premiumYearlyPlan,
}

var ftcPlansSandbox = buildSandboxPlans()

// Index FTC plans by stripe plan id.
var ftcPlansByStripeIDLive = ftcPlansLive.WithStripeIndex()

var ftcPlansByStripeIDTest = buildSandboxPlans().WithStripeIndex()

func GetFtcPlans(live bool) FtcPlans {
	if live {
		return ftcPlansLive
	}

	return ftcPlansSandbox
}

func GetFtcPlansWithStripe(live bool) FtcPlans {
	if live {
		return ftcPlansByStripeIDLive
	}

	return ftcPlansByStripeIDTest
}
