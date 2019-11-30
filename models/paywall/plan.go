package paywall

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
	"math"
	"strings"
	"time"
)

var (
	standardMonthlyPlan = Plan{
		Coordinate: Coordinate{
			Tier:       enum.TierStandard,
			Cycle:      enum.CycleMonth,
			LegacyTier: null.IntFrom(5),
		},
		ListPrice:        28.00,
		NetPrice:         28.00,
		Title:            "FT中文网 - 月度标准会员",
		CycleCount:       1,
		Currency:         "cny",
		ExtraDays:        1,
		stripeLivePlanID: "plan_FXZYLOEbcvj5Tx",
		stripeTestPlanID: "plan_FOdgPTznDwHU4i",
		AppleProductID:   "com.ft.ftchinese.mobile.subscription.member.monthly",
	}

	standardYearlyPlan = Plan{
		Coordinate: Coordinate{
			Tier:       enum.TierStandard,
			Cycle:      enum.CycleYear,
			LegacyTier: null.IntFrom(10),
		},
		ListPrice:        258.00,
		NetPrice:         258.00,
		Title:            "FT中文网 - 年度标准会员",
		CycleCount:       1,
		Currency:         "cny",
		ExtraDays:        1,
		stripeLivePlanID: "plan_FXZZUEDpToPlZK",
		stripeTestPlanID: "plan_FOdfeaqzczp6Ag",
		AppleProductID:   "com.ft.ftchinese.mobile.subscription.member",
	}

	premiumYearlyPlan = Plan{
		Coordinate: Coordinate{
			Tier:       enum.TierPremium,
			Cycle:      enum.CycleYear,
			LegacyTier: null.IntFrom(100),
		},
		ListPrice:        1998.00,
		NetPrice:         1998.00,
		Title:            "FT中文网 - 年度高端会员",
		CycleCount:       1,
		Currency:         "cny",
		ExtraDays:        1,
		stripeLivePlanID: "plan_FXZbv1cDTsUKOg",
		stripeTestPlanID: "plan_FOde0uAr0V4WmT",
		AppleProductID:   "com.ft.ftchinese.mobile.subscription.vip",
	}
)

var ordinals = map[string]int{
	"standard_month": 5,
	"standard_year":  10,
	"premium_year":   100,
}

// Coordinate includes a product and a plan billing cycle to identify the plan to subscribe.
type Coordinate struct {
	Tier       enum.Tier  `json:"tier" db:"sub_tier"`
	Cycle      enum.Cycle `json:"cycle" db:"sub_cycle"`
	LegacyTier null.Int   `json:"-" db:"vip_type"`
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
	ListPrice        float64 `json:"listPrice" db:"price"`
	NetPrice         float64 `json:"netPrice" db:"amount"`
	CycleCount       int64   `json:"cycleCount" db:"cycle_count"`
	ExtraDays        int64   `json:"extraDays" db:"extra_days"`
	Currency         string  `json:"currency" db:"currency"`
	Title            string  `json:"description" db:"title"`
	stripeLivePlanID string  `json:"-"`
	stripeTestPlanID string  `json:"-"`
	AppleProductID   string  `json:"-"`
}

func (p Plan) GetStripePlanID(live bool) string {
	if live {
		return p.stripeLivePlanID
	}

	return p.stripeTestPlanID
}

// withSandboxPrice returns the sandbox version of a plan.
func (p Plan) withSandboxPrice() Plan {
	p.NetPrice = 0.01
	return p
}

func (p Plan) WithStripePrice(sp stripe.Plan) Plan {
	p.ListPrice = float64(sp.Amount / 100)
	p.NetPrice = p.ListPrice
	p.Currency = string(sp.Currency)

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
	p.Title = "FT中文网 - 升级高端会员"

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

func (p Plan) GetPeriodEnd(start time.Time) (time.Time, error) {

	switch p.Cycle {
	case enum.CycleYear:
		return start.AddDate(int(p.CycleCount), 0, int(p.ExtraDays)), nil

	case enum.CycleMonth:
		return start.AddDate(0, int(p.CycleCount), int(p.ExtraDays)), nil

	default:
		return time.Time{}, errors.New("invalid plan cycle")
	}
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
