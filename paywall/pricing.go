package paywall

import (
	"fmt"
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
