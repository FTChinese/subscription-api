package shared

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/patrickmn/go-cache"
)

type PaywallCommon struct {
	dbs   db.ReadWriteMyDBs
	cache *cache.Cache
}

func NewPaywallCommon(dbs db.ReadWriteMyDBs, c *cache.Cache) PaywallCommon {
	return PaywallCommon{
		dbs:   dbs,
		cache: c,
	}
}

func (env PaywallCommon) ClearCache() {
	env.cache.Flush()
}

// LoadPaywall tries to load paywall from cache.
// Fallback to db if not found in cache.
func (env PaywallCommon) LoadPaywall(live bool) (pw.Paywall, error) {
	x, found := env.cache.Get(ids.PaywallCacheKey(live))

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
func (env PaywallCommon) retrievePaywall(live bool) (pw.Paywall, error) {
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
func (env PaywallCommon) cachePaywall(p pw.Paywall) {
	env.cache.Set(
		ids.PaywallCacheKey(p.LiveMode),
		p,
		cache.NoExpiration)
}

// RetrievePaywallDoc loads the latest row of paywall document.
func (env PaywallCommon) RetrievePaywallDoc(live bool) (pw.PaywallDoc, error) {
	var pwb pw.PaywallDoc

	err := env.dbs.Read.Get(
		&pwb,
		pw.StmtRetrievePaywallDoc,
		live)

	if err != nil {
		if err != sql.ErrNoRows {
			return pw.PaywallDoc{}, err
		}

		// No paywall doc exists yet. Returns an empty version.
		return pw.PaywallDoc{
			LiveMode: live,
		}, nil
	}

	return pwb, nil
}

type pwDocResult struct {
	value pw.PaywallDoc
	error error
}

// asyncPwDoc loads paywall document in background.
func (env PaywallCommon) asyncPwDoc(live bool) <-chan pwDocResult {
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
func (env PaywallCommon) retrieveActiveProducts(live bool) ([]pw.Product, error) {
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
func (env PaywallCommon) asyncRetrieveActiveProducts(live bool) <-chan productsResult {
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
	value []pw.PaywallPrice
	error error
}

// ListActivePrices lists active prices of products on paywall, directly from DB.
func (env PaywallCommon) ListActivePrices(live bool) ([]pw.PaywallPrice, error) {
	var prices = make([]pw.PaywallPrice, 0)

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
func (env PaywallCommon) asyncListActivePrices(live bool) <-chan activePricesResult {
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
