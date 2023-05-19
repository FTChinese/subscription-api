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

// DeactivateFtcSiblingPrice turns to false the is_active flag of
// all prices' under a product of the same edition
// to ensure there won't be multiple editions being in active
// state simultaneously.
// The passed-in PaywallPrice is the one that should be excluded
// from deactivating.s
func (tx PriceTx) DeactivateFtcSiblingPrice(p price.FtcPrice) error {
	_, err := tx.NamedExec(
		price.StmtDeactivateSiblingPrices,
		p)

	if err != nil {
		return err
	}

	return nil
}

// ActivateFtcPrice turns the is_active state to true for the
// specified price.
// It should be called immediately following the above one.
func (tx PriceTx) ActivateFtcPrice(p price.FtcPrice) error {
	_, err := tx.NamedExec(price.StmtFtcPriceState, p)
	if err != nil {
		return err
	}

	return nil
}

// UpsertActivePrice inserts a new entry into subs_product.product_active_plans,
// or update it in case of id conflict.
func (tx PriceTx) UpsertActivePrice(p price.ActivePrice) error {
	_, err := tx.NamedExec(
		price.StmtUpsertActivePrice,
		p)

	if err != nil {
		return err
	}

	return nil
}

func (tx PriceTx) DeactivateFtcPrice(p price.FtcPrice) error {
	_, err := tx.NamedExec(
		price.StmtFtcPriceState,
		p)

	if err != nil {
		return err
	}

	return nil
}

// RemoveFtcActivePrice from product_active_price if
// the table has this entry.
// DO NOT use the id field to perform this operation
// since you might be accidentally remove a valid
// price in case the the price to be removed
// is attached to an inactive product.
func (tx PriceTx) RemoveFtcActivePrice(p price.FtcPrice) error {
	_, err := tx.Exec(
		price.StmtRemoveFtcActivePrice,
		p.ID)
	if err != nil {
		return err
	}

	return nil
}
