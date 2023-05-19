package txrepo

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/jmoiron/sqlx"
)

type ProductTx struct {
	*sqlx.Tx
}

func BeginProductTx(dbx *sqlx.DB) (ProductTx, error) {
	tx, err := dbx.Beginx()
	if err != nil {
		return ProductTx{}, err
	}

	return ProductTx{
		Tx: tx,
	}, nil
}

// DeactivateSiblingProduct turns products of same tier in same mode to inactive except the specified one.
func (tx ProductTx) DeactivateSiblingProduct(p reader.Product) error {
	_, err := tx.NamedExec(
		reader.StmtDeactivateSiblingProducts,
		p)

	if err != nil {
		return err
	}

	return nil
}

func (tx ProductTx) ActivateProduct(p reader.Product) error {
	_, err := tx.NamedExec(
		reader.StmtActivateProduct,
		p)

	if err != nil {
		return err
	}

	return nil
}

func (tx ProductTx) SetProductOnPaywallLegacy(p reader.Product) error {
	_, err := tx.NamedExec(
		reader.StmtPutProductOnPaywallLegacy,
		p)

	if err != nil {
		return err
	}

	return nil
}

func (tx ProductTx) SetProductOnPaywall(p reader.Product) error {
	_, err := tx.NamedExec(
		reader.StmtPutProductOnPaywall,
		p)

	if err != nil {
		return err
	}

	return nil
}
