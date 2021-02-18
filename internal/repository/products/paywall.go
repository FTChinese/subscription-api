package products

import (
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/patrickmn/go-cache"
)

// retrieveBanner retrieves a banner and the optional promo attached to it.
// The banner id is fixed to 1.
func (env Env) retrieveBanner() (product.BannerSchema, error) {
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

// asyncLoadBanner retrieves banner in a goroutine.
func (env Env) asyncLoadBanner() <-chan bannerResult {
	c := make(chan bannerResult)

	go func() {
		defer close(c)

		pw, err := env.retrieveBanner()

		c <- bannerResult{
			value: pw,
			error: err,
		}
	}()

	return c
}

// retrievePaywall retrieves all elements of paywall concurrently
// and then build them into a single Paywall instance.
func (env Env) retrievePaywall() (product.Paywall, error) {
	bannerCh, productsCh, plansCh := env.asyncLoadBanner(), env.asyncRetrieveProducts(), env.asyncPricesFromDB()

	// Retrieve banner and its promo, products, and each product's plans
	// in 3 goroutine.
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

	// Zip products with its plans.
	products := product.BuildPaywallProducts(productsRes.value, plansRes.value)

	// Build paywall.
	return product.NewPaywall(bannerRes.value, products), nil
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

	// If found in cache and it can be casted to Paywall, return it;
	// otherwise retrieve from DB.
	if found {
		if pw, ok := x.(product.Paywall); ok {
			return pw, nil
		}
	}

	pw, err := env.retrievePaywall()
	if err != nil {
		return product.Paywall{}, err
	}

	// Cache it.
	env.cachePaywall(pw)

	return pw, nil
}
