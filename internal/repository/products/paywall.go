package products

import (
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/patrickmn/go-cache"
)

// retrieveBanner retrieves a banner and the optional promo attached to it.
// The banner id is fixed to 1.
func (env Env) retrieveBanner() (pw.BannerSchema, error) {
	var schema pw.BannerSchema

	err := env.dbs.Read.Get(&schema, pw.StmtBanner)
	if err != nil {
		return pw.BannerSchema{}, err
	}

	return schema, nil
}

type bannerResult struct {
	value pw.BannerSchema
	error error
}

// asyncRetrieveBanner retrieves banner in a goroutine.
func (env Env) asyncRetrieveBanner() <-chan bannerResult {
	c := make(chan bannerResult)

	go func() {
		defer close(c)

		banner, err := env.retrieveBanner()

		c <- bannerResult{
			value: banner,
			error: err,
		}
	}()

	return c
}

// retrievePaywall retrieves all elements of paywall concurrently
// and then build them into a single Paywall instance.
func (env Env) retrievePaywall() (pw.Paywall, error) {
	bannerCh, productsCh, plansCh := env.asyncRetrieveBanner(), env.asyncRetrieveProducts(), env.asyncRetrieveProductPrices()

	// Retrieve banner and its promo, products, and each price's plans
	// in 3 goroutine.
	bannerRes, productsRes, plansRes := <-bannerCh, <-productsCh, <-plansCh

	if bannerRes.error != nil {
		return pw.Paywall{}, bannerRes.error
	}

	if productsRes.error != nil {
		return pw.Paywall{}, productsRes.error
	}

	if plansRes.error != nil {
		return pw.Paywall{}, plansRes.error
	}

	// Zip products with its plans.
	products := pw.NewPaywallProducts(productsRes.value, plansRes.value)

	// Build paywall.
	return pw.NewPaywall(bannerRes.value, products), nil
}

// cachePaywall caches paywall data after retrieved from db.
func (env Env) cachePaywall(p pw.Paywall) {
	env.cache.Set(keyPaywall, p, cache.NoExpiration)
}

func (env Env) ClearCache() {
	env.cache.Flush()
}

// LoadPaywall tries to load paywall from cache.
// Fallback to db if not found in cache.
func (env Env) LoadPaywall() (pw.Paywall, error) {
	x, found := env.cache.Get(keyPaywall)

	// If found in cache and it can be casted to Paywall, return it;
	// otherwise retrieve from DB.
	if found {
		if paywall, ok := x.(pw.Paywall); ok {
			return paywall, nil
		}
	}

	paywall, err := env.retrievePaywall()
	if err != nil {
		return pw.Paywall{}, err
	}

	// Cache it.
	env.cachePaywall(paywall)

	return paywall, nil
}
