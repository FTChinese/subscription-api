package repository

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

type PaywallRepo struct {
	dbs db.ReadWriteMyDBs
}

func NewPaywallRepo(dbs db.ReadWriteMyDBs) PaywallRepo {
	return PaywallRepo{
		dbs: dbs,
	}
}

// RetrievePaywall retrieves all elements of paywall concurrently
// and then build them into a single Paywall instance.
func (repo PaywallRepo) RetrievePaywall(live bool) (reader.Paywall, error) {
	pwDocCh, productsCh, plansCh := repo.asyncPwDoc(live), repo.asyncRetrieveActiveProducts(live), repo.asyncListActivePrices(live)

	// Retrieve banner and its promo, products, and each price's plans
	// in 3 goroutine.
	pwDocRes, productsRes, plansRes := <-pwDocCh, <-productsCh, <-plansCh

	if pwDocRes.error != nil {
		return reader.Paywall{}, pwDocRes.error
	}

	if productsRes.error != nil {
		return reader.Paywall{}, productsRes.error
	}

	if plansRes.error != nil {
		return reader.Paywall{}, plansRes.error
	}

	// Zip products with its plans.
	products := reader.NewPaywallProducts(productsRes.value, plansRes.value)

	// Build paywall.
	return reader.NewPaywall(pwDocRes.value, products), nil
}

// RetrievePaywallDoc loads the latest row of paywall document.
func (repo PaywallRepo) RetrievePaywallDoc(live bool) (reader.PaywallDoc, error) {
	var pwb reader.PaywallDoc

	err := repo.dbs.Read.Get(
		&pwb,
		reader.StmtRetrievePaywallDoc,
		live)

	if err != nil {
		if err != sql.ErrNoRows {
			return reader.PaywallDoc{}, err
		}

		// No paywall doc exists yet. Returns an empty version.
		return reader.PaywallDoc{
			LiveMode: live,
		}, nil
	}

	return pwb, nil
}

type pwDocResult struct {
	value reader.PaywallDoc
	error error
}

// asyncPwDoc loads paywall document in background.
func (repo PaywallRepo) asyncPwDoc(live bool) <-chan pwDocResult {
	c := make(chan pwDocResult)

	go func() {
		defer close(c)

		pwDoc, err := repo.RetrievePaywallDoc(live)

		c <- pwDocResult{
			value: pwDoc,
			error: err,
		}
	}()

	return c
}

type productsResult struct {
	value []reader.Product
	error error
}

// retrieveActiveProducts retrieve all products present on paywall.
func (repo PaywallRepo) retrieveActiveProducts(live bool) ([]reader.Product, error) {
	var products = make([]reader.Product, 0)

	err := repo.dbs.Read.Select(
		&products,
		reader.StmtPaywallProducts,
		live)

	if err != nil {
		return nil, err
	}

	return products, nil
}

// asyncRetrieveActiveProducts retrieves products in a goroutine.
func (repo PaywallRepo) asyncRetrieveActiveProducts(live bool) <-chan productsResult {
	ch := make(chan productsResult)

	go func() {
		products, err := repo.retrieveActiveProducts(live)

		ch <- productsResult{
			value: products,
			error: err,
		}
	}()

	return ch
}

// activePricesResult contains a list of pricing plans and error occurred.
type activePricesResult struct {
	value []reader.PaywallPrice
	error error
}

// ListActivePrices lists active prices of products on paywall, directly from DB.
func (repo PaywallRepo) ListActivePrices(live bool) ([]reader.PaywallPrice, error) {
	var prices = make([]reader.PaywallPrice, 0)

	err := repo.dbs.Read.Select(
		&prices,
		reader.StmtListPaywallPrice,
		live)
	if err != nil {
		return nil, err
	}

	return prices, nil
}

// asyncListActivePrices retrieves a list of plans in a goroutine.
// This is used to construct the paywall data.
func (repo PaywallRepo) asyncListActivePrices(live bool) <-chan activePricesResult {
	ch := make(chan activePricesResult)

	go func() {
		defer close(ch)

		plans, err := repo.ListActivePrices(live)

		ch <- activePricesResult{
			value: plans,
			error: err,
		}
	}()

	return ch
}
