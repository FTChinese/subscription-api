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

// DeactivateSiblingPrice turns to false the is_active flag of
// all prices' under a product of the same edition
// to ensure there won't be multiple editions being in active
// state simultaneously.
// The passed-in PaywallPrice is the one that should be excluded
// from deactivating.s
func (tx PriceTx) DeactivateSiblingPrice(p price.FtcPrice) error {
	_, err := tx.NamedExec(
		price.StmtDeactivateSiblingPrices,
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
