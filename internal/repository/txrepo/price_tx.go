package txrepo

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/jmoiron/sqlx"
)

type PriceTx struct {
	*sqlx.Tx
}

func NewPriceTx(tx *sqlx.Tx) PriceTx {
	return PriceTx{
		Tx: tx,
	}
}

// RetrieveFtcPrice locks a row in price and retrieve it.
func (tx PriceTx) RetrieveFtcPrice(id string) (price.FtcPrice, error) {
	var ftcPrice price.FtcPrice
	err := tx.Get(&ftcPrice, price.StmtLockFtcPrice, id)
	if err != nil {
		return price.FtcPrice{}, err
	}

	return ftcPrice, nil
}

// DeactivateSiblingPrice turns to false the is_active flag of
// all prices' under a product of the same edition
// to ensure there won't be multiple editions being in active
// state simultaneously.
// The passed-in FtcPrice is the one that should be excluded
// from deactivating.s
func (tx PriceTx) DeactivateSiblingPrice(p price.FtcPrice) error {
	_, err := tx.NamedExec(
		price.StmtDeactivatePricesOfSameEdition,
		p)

	if err != nil {
		return err
	}

	return nil
}

// ActivatePrice turns the is_active state to true for the
// specified price.
// It should be called immediately following the above one.
func (tx PriceTx) ActivatePrice(p price.FtcPrice) error {
	_, err := tx.NamedExec(price.StmtActivatePrice, p)
	if err != nil {
		return err
	}

	return nil
}

func (tx PriceTx) ActivatePriceLegacy(p price.FtcPrice) error {
	_, err := tx.NamedExec(price.StmtActivatePriceUnderProduct, p)
	if err != nil {
		return err
	}

	return nil
}
