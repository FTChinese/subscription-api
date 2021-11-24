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
	pwDocCh, productsCh, plansCh := env.asyncPwDoc(live), env.asyncRetrieveActiveProducts(live), env.asyncListActivePrices(live)

	// Retrieve banner and its promo, products, and each price's plans
	// in 3 goroutine.
	pwDocRes, productsRes, plansRes := <-pwDocCh, <-productsCh, <-plansCh

	if pwDocRes.error != nil {
		return pw.Paywall{}, pwDocRes.error
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
	return pw.NewPaywall(pwDocRes.value, products), nil
}

// cachePaywall caches paywall data after retrieved from db.
func (env Env) cachePaywall(p pw.Paywall) {
	env.cache.Set(
		getPaywallCacheKey(p.LiveMode),
		p,
		cache.NoExpiration)
}

type pwDocResult struct {
	value pw.PaywallDoc
	error error
}

func (env Env) asyncPwDoc(live bool) <-chan pwDocResult {
	c := make(chan pwDocResult)

	go func() {
		defer close(c)

		pwDoc, err := env.RetrievePaywallDoc(live)

		c <- pwDocResult{
			value: pwDoc,
			error: err,
		}
	}()

	return c
}

type productsResult struct {
	value []pw.Product
	error error
}

// retrieveActiveProducts retrieve all products present on paywall.
func (env Env) retrieveActiveProducts(live bool) ([]pw.Product, error) {
	var products = make([]pw.Product, 0)

	err := env.dbs.Read.Select(
		&products,
		pw.StmtPaywallProducts,
		live)

	if err != nil {
		return nil, err
	}

	return products, nil
}

// asyncRetrieveActiveProducts retrieves products in a goroutine.
func (env Env) asyncRetrieveActiveProducts(live bool) <-chan productsResult {
	ch := make(chan productsResult)

	go func() {
		products, err := env.retrieveActiveProducts(live)

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

// ListActivePrices lists active prices of products on paywall, directly from DB.
func (env Env) ListActivePrices(live bool) ([]price.FtcPrice, error) {
	var prices = make([]price.FtcPrice, 0)

	err := env.dbs.Read.Select(
		&prices,
		price.StmtListPaywallPrice,
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

		plans, err := env.ListActivePrices(live)

		ch <- activePricesResult{
			value: plans,
			error: err,
		}
	}()

	return ch
}

func (env Env) ClearCache() {
	env.cache.Flush()
}
