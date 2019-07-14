package paywall

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
)

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

// FtcPlans defines a collection pricing plan.
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
