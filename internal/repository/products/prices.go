package products

import (
	"github.com/FTChinese/subscription-api/pkg/price"
)

// CreatePrice inserts a row into plan table.
func (env Env) CreatePrice(p price.FtcPrice) error {
	_, err := env.dbs.Write.NamedExec(price.StmtCreatePrice, p)
	if err != nil {
		return err
	}

	return nil
}

// ActivatePrice flags a price as active while all other
// prices of the same edition and same live mode under the same product id
// is turned to inactive.
func (env Env) ActivatePrice(id string) (price.FtcPrice, error) {
	tx, err := env.dbs.Write.Beginx()
	if err != nil {
		return price.FtcPrice{}, err
	}

	// Retrieve the price to activate.
	var ftcPrice price.FtcPrice
	err = tx.Get(&ftcPrice, price.StmtLockFtcPrice, id)
	if err != nil {
		_ = tx.Rollback()
		return price.FtcPrice{}, err
	}

	// Deactivate all other prices.
	_, err = tx.NamedExec(price.StmtDeactivatePricesOfSameEdition, ftcPrice)
	if err != nil {
		_ = tx.Rollback()
		return price.FtcPrice{}, err
	}

	ftcPrice = ftcPrice.Activate()
	// Activate the price
	_, err = tx.NamedExec(price.StmtActivatePrice, ftcPrice)
	if err != nil {
		_ = tx.Rollback()
		return price.FtcPrice{}, err
	}

	// Handle legacy activation approach.
	_, err = tx.NamedExec(price.StmtActivatePriceLegacy, ftcPrice)
	if err != nil {
		_ = tx.Rollback()
		return price.FtcPrice{}, err
	}

	if err := tx.Commit(); err != nil {
		return price.FtcPrice{}, err
	}

	return ftcPrice, nil
}

// RetrieveFtcPrice retrieves a single row by plan id.
func (env Env) RetrieveFtcPrice(id string) (price.FtcPrice, error) {
	var p price.FtcPrice
	err := env.dbs.Read.Get(&p, price.StmtFtcPrice, id)
	if err != nil {
		return price.FtcPrice{}, err
	}

	return p, nil
}

// UpdateFtcPriceOffers after a new discount is created/paused/cancelled under this price.
func (env Env) UpdateFtcPriceOffers(f price.FtcPrice) error {
	_, err := env.dbs.Write.NamedExec(price.StmtSetPriceOffers, f)

	if err != nil {
		return err
	}

	return nil
}

// RefreshFtcPriceOffers updated a price's discount list.
func (env Env) RefreshFtcPriceOffers(f price.FtcPrice) (price.FtcPrice, error) {
	offers, err := env.ListActiveDiscounts(f.ID, f.LiveMode)
	if err != nil {
		return price.FtcPrice{}, err
	}

	updated := f.SetOffers(offers)
	err = env.UpdateFtcPriceOffers(updated)
	if err != nil {
		return price.FtcPrice{}, err
	}

	return updated, nil
}

// ListPrices retrieves all prices of a product, regardless whether they are live or not.
// This is used by CMS to list a product's prices so that
// user should be able to activate an inactive one.
func (env Env) ListPrices(prodID string) ([]price.FtcPrice, error) {
	var list = make([]price.FtcPrice, 0)
	err := env.dbs.Read.Select(
		&list,
		price.StmtListPricesOfProduct,
		prodID)

	if err != nil {
		return nil, err
	}

	return list, nil
}
