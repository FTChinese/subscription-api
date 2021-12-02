package products

import (
	"github.com/FTChinese/subscription-api/pkg/price"
)

// CreatePrice inserts a row into plan table.
func (env Env) CreatePrice(p price.FtcPrice) error {
	_, err := env.DBs.Write.NamedExec(price.StmtCreatePrice, p)
	if err != nil {
		return err
	}

	return nil
}

// UpdateFtcPrice updates a price's description and stripe price id.
func (env Env) UpdateFtcPrice(f price.FtcPrice) error {
	_, err := env.DBs.Write.NamedExec(
		price.StmtUpdatePrice,
		f)

	if err != nil {
		return err
	}

	return nil
}

// ActivatePrice flags a price as active while all other
// prices of the same edition and same live mode under the same product id
// is turned to inactive.
func (env Env) ActivatePrice(ftcPrice price.FtcPrice) error {
	tx, err := env.beginPriceTx()
	if err != nil {
		return err
	}

	// Deactivate all other prices.
	err = tx.DeactivateSiblingPrice(ftcPrice)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	ftcPrice = ftcPrice.Activate()

	// Activate the price
	err = tx.ActivatePrice(ftcPrice)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Handle legacy activation approach.
	// Due to table's unique constraint design, product_id and cycle combined must be unique,
	// we cannot insert multiple price of the same cycle under the same product.
	if ftcPrice.LiveMode {
		err = tx.ActivatePriceLegacy(ftcPrice)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// UpdateFtcPriceOffers after a new discount is created/paused/cancelled under this price.
func (env Env) UpdateFtcPriceOffers(f price.FtcPrice) error {
	_, err := env.DBs.Write.NamedExec(price.StmtSetPriceOffers, f)

	if err != nil {
		return err
	}

	return nil
}

// RefreshFtcPriceOffers retrieves all discounts of a price
// and save them as JSON in the price's row.
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
func (env Env) ListPrices(prodID string, live bool) ([]price.FtcPrice, error) {
	var list = make([]price.FtcPrice, 0)
	err := env.DBs.Read.Select(
		&list,
		price.StmtListPricesOfProduct,
		prodID,
		live)

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (env Env) ArchivePrice(p price.FtcPrice) error {
	_, err := env.DBs.Write.NamedExec(price.StmtArchivePrice, p)
	if err != nil {
		return err
	}

	return nil
}

// ArchivePriceDiscounts turns all discount under a price into
// cancelled mode.
func (env Env) ArchivePriceDiscounts(p price.FtcPrice) error {
	_, err := env.DBs.Write.NamedExec(price.StmtArchivePriceDiscounts, p)
	if err != nil {
		return err
	}

	return nil
}
