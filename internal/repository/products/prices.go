package products

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/patrickmn/go-cache"
)

// retrieveProductPrices lists active product prices on paywall, directly from DB.
func (env Env) retrieveProductPrices() ([]price.FtcPrice, error) {
	var schema = make([]pw.PriceSchema, 0)
	var prices = make([]price.FtcPrice, 0)

	err := env.db.Select(&schema, pw.StmtActiveProductPrices)
	if err != nil {
		return nil, err
	}

	for _, v := range schema {
		prices = append(prices, pw.NewFtcPrice(v))
	}

	return prices, nil
}

// pricesResult contains a list of pricing plans and error occurred.
type pricesResult struct {
	value []price.FtcPrice
	error error
}

// asyncRetrieveProductPrices retrieves a list of plans in a goroutine.
// This is used to construct the paywall data.
func (env Env) asyncRetrieveProductPrices() <-chan pricesResult {
	ch := make(chan pricesResult)

	go func() {
		defer close(ch)

		plans, err := env.retrieveProductPrices()

		ch <- pricesResult{
			value: plans,
			error: err,
		}
	}()

	return ch
}

// cacheActivePrices caching all currently active prices as an array.
func (env Env) cacheActivePrices(p []price.FtcPrice) {
	env.cache.Set(keyPricing, p, cache.DefaultExpiration)
}

// ActivePricesFromCacheOrDB tries to load all active pricing plans from cache,
// then fallback to db if not found. If retrieved from DB,
// the data will be cached.
func (env Env) ActivePricesFromCacheOrDB() ([]price.FtcPrice, error) {
	x, found := env.cache.Get(keyPricing)

	if found {
		if p, ok := x.([]price.FtcPrice); ok {
			return p, nil
		}
	}

	p, err := env.retrieveProductPrices()
	if err != nil {
		return nil, err
	}

	env.cacheActivePrices(p)

	return p, nil
}

// RetrievePrice retrieves a plan with discount by ID.
func (env Env) RetrievePrice(id string) (price.FtcPrice, error) {
	var schema pw.PriceSchema

	err := env.db.Get(&schema, pw.StmtProductPrice, id)
	if err != nil {
		return price.FtcPrice{}, nil
	}

	return pw.NewFtcPrice(schema), nil
}

// ActivePriceOfEdition retrieves an active plan by Edition.
func (env Env) ActivePriceOfEdition(e price.Edition) (price.FtcPrice, error) {
	prices, err := env.ActivePricesFromCacheOrDB()
	if err != nil {
		return price.FtcPrice{}, err
	}

	for _, v := range prices {
		if v.Edition == e {
			return v, nil
		}
	}

	return price.FtcPrice{}, sql.ErrNoRows
}
