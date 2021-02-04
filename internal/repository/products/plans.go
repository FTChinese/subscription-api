package products

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/patrickmn/go-cache"
)

// listActivePlans retrieves list plans present on paywall, directly from DB.
func (env Env) listActivePlans() ([]product.ExpandedPlan, error) {
	var schema = make([]product.ExpandedPlanSchema, 0)
	var plans = make([]product.ExpandedPlan, 0)

	err := env.db.Select(&schema, product.StmtActivePlans)
	if err != nil {
		return nil, err
	}

	for _, v := range schema {
		plans = append(plans, v.ExpandedPlan())
	}

	return plans, nil
}

// cacheActivePlans caching all currently active plans.
func (env Env) cacheActivePlans(p []product.ExpandedPlan) {
	env.cache.Set(keyPricing, p, cache.DefaultExpiration)
}

// ListActivePlans tries to load all active pricing plans from cache,
// then fallback to db if not found. If retrieved from DB,
// the data will be cached.
func (env Env) ListActivePlans() ([]product.ExpandedPlan, error) {
	x, found := env.cache.Get(keyPricing)

	if found {
		if p, ok := x.([]product.ExpandedPlan); ok {
			return p, nil
		}
	}

	p, err := env.listActivePlans()
	if err != nil {
		return nil, err
	}

	env.cacheActivePlans(p)

	return p, nil
}

// RetrievePlan retrieves a plan with discount by ID.
func (env Env) RetrievePlan(id string) (product.ExpandedPlan, error) {
	var schema product.ExpandedPlanSchema

	err := env.db.Get(&schema, product.StmtExpandedPlan, id)
	if err != nil {
		return product.ExpandedPlan{}, nil
	}

	return schema.ExpandedPlan(), nil
}

// FindActivePlan retrieves an active plan by tier and cycle.
func (env Env) FindActivePlan(e product.Edition) (product.ExpandedPlan, error) {
	plans, err := env.ListActivePlans()
	if err != nil {
		return product.ExpandedPlan{}, err
	}

	for _, v := range plans {
		if v.Edition == e {
			return v, nil
		}
	}

	return product.ExpandedPlan{}, sql.ErrNoRows
}

// plansResult contains a list of pricing plans and error occurred.
type plansResult struct {
	value []product.ExpandedPlan
	error error
}

// asyncLoadPlans retrieves a list of plans in a goroutine.
func (env Env) asyncLoadPlans() <-chan plansResult {
	ch := make(chan plansResult)

	go func() {
		defer close(ch)

		plans, err := env.listActivePlans()

		ch <- plansResult{
			value: plans,
			error: err,
		}
	}()

	return ch
}
