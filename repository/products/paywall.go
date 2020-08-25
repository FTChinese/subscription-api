package products

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/patrickmn/go-cache"
)

// loadBanner retrieves a banner and promo. The banner id is fixed to 1.
func (env Env) loadBanner() (product.BannerSchema, error) {
	var schema product.BannerSchema

	err := env.db.Get(&schema, product.StmtBanner)
	if err != nil {
		return product.BannerSchema{}, err
	}

	return schema, nil
}

type bannerResult struct {
	value product.BannerSchema
	error error
}

func (env Env) asyncLoadBanner() <-chan bannerResult {
	c := make(chan bannerResult)

	go func() {
		defer close(c)

		pw, err := env.loadBanner()

		c <- bannerResult{
			value: pw,
			error: err,
		}
	}()

	return c
}

// LoadPaywallProducts retrieve all products present on paywall.
func (env Env) loadProducts() ([]product.Product, error) {
	var products = make([]product.Product, 0)

	err := env.db.Select(&products, product.StmtPaywallProducts)

	if err != nil {
		return nil, err
	}

	return products, nil
}

type productsResult struct {
	value []product.Product
	error error
}

func (env Env) asyncLoadProducts() <-chan productsResult {
	ch := make(chan productsResult)

	go func() {
		products, err := env.loadProducts()

		ch <- productsResult{
			value: products,
			error: err,
		}
	}()

	return ch
}

func (env Env) retrievePaywall() (product.Paywall, error) {
	bannerCh, productsCh, plansCh := env.asyncLoadBanner(), env.asyncLoadProducts(), env.asyncLoadPlans()

	bannerRes, productsRes, plansRes := <-bannerCh, <-productsCh, <-plansCh

	if bannerRes.error != nil {
		return product.Paywall{}, bannerRes.error
	}

	if productsRes.error != nil {
		return product.Paywall{}, productsRes.error
	}

	if plansRes.error != nil {
		return product.Paywall{}, plansRes.error
	}

	products := product.BuildPaywallProducts(productsRes.value, plansRes.value)

	paywall := product.NewPaywall(bannerRes.value, products)

	env.cachePaywall(paywall)

	return paywall, nil
}

// cachePaywall caches paywall data after retrieved from db.
func (env Env) cachePaywall(p product.Paywall) {
	env.cache.Set(keyPaywall, p, cache.NoExpiration)
}

func (env Env) ClearCache() {
	env.cache.Flush()
}

// LoadPaywall tries to load paywall from cache.
// Fallback to db if not found in cache.
func (env Env) LoadPaywall() (product.Paywall, error) {
	x, found := env.cache.Get(keyPaywall)

	if !found {
		return env.retrievePaywall()
	}

	if pw, ok := x.(product.Paywall); ok {
		return pw, nil
	}

	return product.Paywall{}, sql.ErrNoRows
}
