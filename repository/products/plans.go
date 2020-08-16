package products

import (
	"database/sql"
	"github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/pkg/product"
)

// loadPlans retrieves all plans present on paywall.
func (env Env) loadPlans() ([]product.ExpandedPlan, error) {
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

type plansResult struct {
	value []product.ExpandedPlan
	error error
}

func (env Env) asyncLoadPlans() <-chan plansResult {
	ch := make(chan plansResult)

	go func() {
		defer close(ch)

		plans, err := env.loadPlans()

		ch <- plansResult{
			value: plans,
			error: err,
		}
	}()

	return ch
}

// LoadPlan retrieves a single plan with discount attached.
func (env Env) retrievePlan(id string) (product.ExpandedPlan, error) {
	var schema product.ExpandedPlanSchema

	err := env.db.Get(&schema, product.StmtExpandedPlan, id)
	if err != nil {
		return product.ExpandedPlan{}, nil
	}

	return schema.ExpandedPlan(), nil
}

func (env Env) cachePlan(p product.ExpandedPlan) {
	env.cache.Set(p.ID, p, cache.DefaultExpiration)
}

func (env Env) LoadPlan(id string) (product.ExpandedPlan, error) {
	x, found := env.cache.Get(id)
	if !found {
		return env.retrievePlan(id)
	}

	if plan, ok := x.(product.ExpandedPlan); ok {
		return plan, nil
	}

	return product.ExpandedPlan{}, sql.ErrNoRows
}
