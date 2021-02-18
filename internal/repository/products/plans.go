package products

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/patrickmn/go-cache"
)

// pricesFromDB retrieves list plans present on paywall, directly from DB.
func (env Env) pricesFromDB() ([]product.Price, error) {
	var schema = make([]product.ExpandedPlanSchema, 0)
	var prices = make([]product.Price, 0)

	err := env.db.Select(&schema, product.StmtActivePlans)
	if err != nil {
		return nil, err
	}

	for _, v := range schema {
		prices = append(prices, product.NewPrice(v))
	}

	return prices, nil
}

// cacheActivePlans caching all currently active plans.
func (env Env) cacheActivePlans(p []product.Price) {
	env.cache.Set(keyPricing, p, cache.DefaultExpiration)
}

// ActivePricesFromCacheOrDB tries to load all active pricing plans from cache,
// then fallback to db if not found. If retrieved from DB,
// the data will be cached.
func (env Env) ActivePricesFromCacheOrDB() ([]product.Price, error) {
	x, found := env.cache.Get(keyPricing)

	if found {
		if p, ok := x.([]product.Price); ok {
			return p, nil
		}
	}

	p, err := env.pricesFromDB()
	if err != nil {
		return nil, err
	}

	env.cacheActivePlans(p)

	return p, nil
}

// RetrievePrice retrieves a plan with discount by ID.
func (env Env) RetrievePrice(id string) (product.Price, error) {
	var schema product.ExpandedPlanSchema

	err := env.db.Get(&schema, product.StmtExpandedPlan, id)
	if err != nil {
		return product.Price{}, nil
	}

	return product.NewPrice(schema), nil
}

// ActivePriceOfEdition retrieves an active plan by Edition.
func (env Env) ActivePriceOfEdition(e product.Edition) (product.Price, error) {
	prices, err := env.ActivePricesFromCacheOrDB()
	if err != nil {
		return product.Price{}, err
	}

	for _, v := range prices {
		if v.Edition == e {
			return v, nil
		}
	}

	return product.Price{}, sql.ErrNoRows
}

// pricesResult contains a list of pricing plans and error occurred.
type pricesResult struct {
	value []product.Price
	error error
}

// asyncPricesFromDB retrieves a list of plans in a goroutine.
// This is used to construct the paywall data.
func (env Env) asyncPricesFromDB() <-chan pricesResult {
	ch := make(chan pricesResult)

	go func() {
		defer close(ch)

		plans, err := env.pricesFromDB()

		ch <- pricesResult{
			value: plans,
			error: err,
		}
	}()

	return ch
}
