package repository

import (
	"database/sql"

	"github.com/FTChinese/subscription-api/pkg"
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

// RetrievePaywall retrieves all components of paywall concurrently
// and then build them into a single Paywall instance.
// TODO: change price retrieval using product_active_price table.
func (repo PaywallRepo) RetrievePaywall(live bool) (reader.Paywall, error) {
	pwDocCh, productsCh, pricesCh := repo.asyncPwDoc(live), repo.asyncRetrieveActiveProducts(live), repo.asyncListActivePrices(live)

	// Retrieve banner and its promo, products, and each price's plans
	// in 3 goroutine.
	pwDocRes, productsRes, pricesRes := <-pwDocCh, <-productsCh, <-pricesCh

	if pwDocRes.Err != nil {
		return reader.Paywall{}, pwDocRes.Err
	}

	if productsRes.Err != nil {
		return reader.Paywall{}, productsRes.Err
	}

	if pricesRes.Err != nil {
		return reader.Paywall{}, pricesRes.Err
	}

	// Zip products with its prices.
	products := reader.NewPaywallProducts(productsRes.Value, pricesRes.Value)

	// Build paywall.
	return reader.NewPaywall(pwDocRes.Value, products).Flatten(), nil
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

// asyncPwDoc loads paywall document in background.
func (repo PaywallRepo) asyncPwDoc(live bool) <-chan pkg.AsyncResult[reader.PaywallDoc] {
	c := make(chan pkg.AsyncResult[reader.PaywallDoc])

	go func() {
		defer close(c)

		pwDoc, err := repo.RetrievePaywallDoc(live)

		c <- pkg.AsyncResult[reader.PaywallDoc]{
			Value: pwDoc,
			Err:   err,
		}
	}()

	return c
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
func (repo PaywallRepo) asyncRetrieveActiveProducts(live bool) <-chan pkg.AsyncResult[[]reader.Product] {
	ch := make(chan pkg.AsyncResult[[]reader.Product])

	go func() {
		products, err := repo.retrieveActiveProducts(live)

		ch <- pkg.AsyncResult[[]reader.Product]{
			Value: products,
			Err:   err,
		}
	}()

	return ch
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
func (repo PaywallRepo) asyncListActivePrices(live bool) <-chan pkg.AsyncResult[[]reader.PaywallPrice] {
	ch := make(chan pkg.AsyncResult[[]reader.PaywallPrice])

	go func() {
		defer close(ch)

		plans, err := repo.ListActivePrices(live)

		ch <- pkg.AsyncResult[[]reader.PaywallPrice]{
			Value: plans,
			Err:   err,
		}
	}()

	return ch
}
