package products

import (
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/patrickmn/go-cache"
)

// retrieveActivePlans retrieves all plans present on paywall, directly from DB.
func (env Env) retrieveActivePlans() ([]product.ExpandedPlan, error) {
	var schema = make([]product.ExpandedPlanSchema, 0)
	var plans = make([]product.ExpandedPlan, 0)

	err := env.db.Select(&schema, product.StmtPaywallPlans)
	if err != nil {
		return nil, err
	}

	for _, v := range schema {
		plans = append(plans, v.ExpandedPlan())
	}

	return plans, nil
}

func (env Env) cachePricing(p []product.ExpandedPlan) {
	env.cache.Set(keyPricing, p, cache.DefaultExpiration)
}

func (env Env) LoadPricing() ([]product.ExpandedPlan, error) {
	x, found := env.cache.Get(keyPricing)

	if found {
		if p, ok := x.([]product.ExpandedPlan); ok {
			return p, nil
		}
	}

	p, err := env.retrieveActivePlans()
	if err != nil {
		return nil, err
	}

	env.cachePricing(p)

	return p, nil
}

type plansResult struct {
	value []product.ExpandedPlan
	error error
}

// asyncLoadPlans retrieves a list of plans in a goroutine.
func (env Env) asyncLoadPlans() <-chan plansResult {
	ch := make(chan plansResult)

	go func() {
		defer close(ch)

		plans, err := env.retrieveActivePlans()

		ch <- plansResult{
			value: plans,
			error: err,
		}
	}()

	return ch
}

// PlanByID retrieves a plan with discount by ID.
func (env Env) PlanByID(id string) (product.ExpandedPlan, error) {
	var schema product.ExpandedPlanSchema

	err := env.db.Get(&schema, product.StmtExpandedPlanByID, id)
	if err != nil {
		return product.ExpandedPlan{}, nil
	}

	return schema.ExpandedPlan(), nil
}

// PlanByEdition retrieves an active plan by tier and cycle.
func (env Env) PlanByEdition(e product.Edition) (product.ExpandedPlan, error) {
	var schema product.ExpandedPlanSchema

	err := env.db.Get(&schema, product.StmtExpandedPlanByEdition, e.Tier, e.Cycle)
	if err != nil {
		return product.ExpandedPlan{}, nil
	}

	return schema.ExpandedPlan(), nil
}
