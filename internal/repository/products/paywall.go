package products

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/patrickmn/go-cache"
)

// LoadPaywall tries to load paywall from cache.
// Fallback to db if not found in cache.
func (env Env) LoadPaywall(live bool) (pw.Paywall, error) {
	x, found := env.cache.Get(getPaywallCacheKey(live))

	// If found in cache, and it can be cast to Paywall, return it;
	// otherwise retrieve from DB.
	if found {
		if paywall, ok := x.(pw.Paywall); ok {
			return paywall, nil
		}
	}

	paywall, err := env.retrievePaywall(live)
	if err != nil {
		return pw.Paywall{}, err
	}

	// Cache it.
	env.cachePaywall(paywall)

	return paywall, nil
}

// retrievePaywall retrieves all elements of paywall concurrently
// and then build them into a single Paywall instance.
func (env Env) retrievePaywall(live bool) (pw.Paywall, error) {
	bannerCh, productsCh, plansCh := env.asyncRetrieveBanner(), env.asyncRetrieveProducts(), env.asyncListActivePrices(live)

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
	return pw.NewPaywall(bannerRes.value, products, live), nil
}

// cachePaywall caches paywall data after retrieved from db.
func (env Env) cachePaywall(p pw.Paywall) {
	env.cache.Set(
		getPaywallCacheKey(p.LiveMode),
		p,
		cache.NoExpiration)
}

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

// retrieveActiveProducts retrieve all products present on paywall.
func (env Env) retrieveActiveProducts() ([]pw.ProductBody, error) {
	var products = make([]pw.ProductBody, 0)

	err := env.dbs.Read.Select(&products, pw.StmtPaywallProducts)

	if err != nil {
		return nil, err
	}

	return products, nil
}

type productsResult struct {
	value []pw.ProductBody
	error error
}

// asyncRetrieveProducts retrieves products in a goroutine.
func (env Env) asyncRetrieveProducts() <-chan productsResult {
	ch := make(chan productsResult)

	go func() {
		products, err := env.retrieveActiveProducts()

		ch <- productsResult{
			value: products,
			error: err,
		}
	}()

	return ch
}

// activePricesResult contains a list of pricing plans and error occurred.
type activePricesResult struct {
	value []price.FtcPrice
	error error
}

// listActivePrices lists active product prices on paywall, directly from DB.
func (env Env) listActivePrices(live bool) ([]price.FtcPrice, error) {
	var prices = make([]price.FtcPrice, 0)

	err := env.dbs.Read.Select(
		&prices,
		price.StmtListActivePrice,
		live)
	if err != nil {
		return nil, err
	}

	return prices, nil
}

// asyncListActivePrices retrieves a list of plans in a goroutine.
// This is used to construct the paywall data.
func (env Env) asyncListActivePrices(live bool) <-chan activePricesResult {
	ch := make(chan activePricesResult)

	go func() {
		defer close(ch)

		plans, err := env.listActivePrices(live)

		ch <- activePricesResult{
			value: plans,
			error: err,
		}
	}()

	return ch
}

// cacheActivePrices caching all currently active prices as an array.
func (env Env) cacheActivePrices(p []price.FtcPrice) {
	if len(p) == 0 {
		return
	}
	env.cache.Set(getActivePricesCacheKey(p[0].LiveMode), p, cache.DefaultExpiration)
}

// ActivePricesFromCacheOrDB tries to load all active pricing plans from cache,
// then fallback to db if not found. If retrieved from DB,
// the data will be cached.
// Deprecated.
func (env Env) ActivePricesFromCacheOrDB(live bool) ([]price.FtcPrice, error) {
	x, found := env.cache.Get(getActivePricesCacheKey(live))

	if found {
		if p, ok := x.([]price.FtcPrice); ok {
			return p, nil
		}
	}

	prices, err := env.listActivePrices(live)
	if err != nil {
		return nil, err
	}

	env.cacheActivePrices(prices)

	return prices, nil
}

func (env Env) ClearCache() {
	env.cache.Flush()
}
