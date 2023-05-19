package txrepo

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/jmoiron/sqlx"
)

type PriceTx struct {
	*sqlx.Tx
}

func BeginPriceTx(dbx *sqlx.DB) (PriceTx, error) {
	tx, err := dbx.Beginx()
	if err != nil {
		return PriceTx{}, err
	}

	return PriceTx{
		Tx: tx,
	}, nil
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

// DeactivateFtcPrice changes a price's is_active or archived fields
func (tx PriceTx) DeactivateFtcPrice(p price.FtcPrice) error {
	_, err := tx.NamedExec(
		price.StmtFtcPriceState,
		p)

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

// RemoveActivePrice from product_active_price if
// the table has this entry.
// DO NOT use the id field to perform this operation
// since you might be accidentally remove a valid
// price in case the the price to be removed
// is attached to an inactive product.
func (tx PriceTx) RemoveActivePrice(p price.ActivePrice) error {
	_, err := tx.NamedExec(
		price.StmtRemoveActivePrice,
		p)
	if err != nil {
		return err
	}

	return nil
}
