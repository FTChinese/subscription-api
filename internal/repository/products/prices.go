package products

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// CreatePrice inserts a row into plan table.
func (env Env) CreatePrice(p price.FtcPrice) error {
	_, err := env.dbs.Write.NamedExec(price.StmtCreatePrice, p)
	if err != nil {
		return err
	}

	return nil
}

func (env Env) CreatePlan(p price.Plan) error {
	_, err := env.dbs.Write.NamedExec(price.StmtCreatePlan, p)
	if err != nil {
		return err
	}

	return nil
}

// UpdatePrice updates a price's description and stripe price id.
func (env Env) UpdatePrice(p price.FtcPrice) error {
	_, err := env.dbs.Write.NamedExec(
		price.StmtUpdatePrice,
		p)

	if err != nil {
		return err
	}

	return nil
}

// ActivatePrice flags a price as active while all other
// prices of the same edition and same live mode under the same product id
// is turned to inactive.
func (env Env) ActivatePrice(p price.FtcPrice) error {
	tx, err := txrepo.BeginPriceTx(env.dbs.Write)
	if err != nil {
		return err
	}

	// Deactivate all other prices.
	err = tx.DeactivateFtcSiblingPrice(p)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Activate the price
	err = tx.ActivateFtcPrice(p)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Also update product_active_price table.
	err = tx.UpsertActivePrice(p.ActiveEntry())
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (env Env) DeactivatePrice(p price.FtcPrice) error {
	tx, err := txrepo.BeginPriceTx(env.dbs.Delete)
	if err != nil {
		return err
	}

	// Flag active to false in price table
	err = tx.DeactivateFtcPrice(p)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Remove the price from product_active_price table if exists.
	err = tx.RemoveActivePrice(p.ActiveEntry())
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// UpdatePriceOffers after a new discount is created/paused/cancelled under this price.
func (env Env) UpdatePriceOffers(pwp reader.PaywallPrice) error {
	_, err := env.dbs.Write.NamedExec(reader.StmtSetPriceOffers, pwp)

	if err != nil {
		return err
	}

	return nil
}

// RefreshPriceOffers retrieves all discounts of a price
// and save them as JSON in the price's row.
func (env Env) RefreshPriceOffers(p reader.PaywallPrice) (reader.PaywallPrice, error) {
	offers, err := env.ListActiveDiscounts(p.ID, p.LiveMode)
	if err != nil {
		return reader.PaywallPrice{}, err
	}

	updated := p.SetOffers(offers)
	err = env.UpdatePriceOffers(updated)
	if err != nil {
		return reader.PaywallPrice{}, err
	}

	return updated, nil
}

// ListProductPrices retrieves all prices of a product, regardless whether they are active or not.
// This is used by CMS to list a product's prices so that
// user should be able to activate an inactive one.
func (env Env) ListProductPrices(prodID string, live bool) ([]reader.PaywallPrice, error) {
	var list = make([]reader.PaywallPrice, 0)
	err := env.dbs.Read.Select(
		&list,
		reader.StmtListProductPrices,
		prodID,
		live)

	if err != nil {
		return nil, err
	}

	return list, nil
}
