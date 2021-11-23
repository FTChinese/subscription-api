package txrepo

import (
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/jmoiron/sqlx"
)

type ProductTx struct {
	*sqlx.Tx
}

func NewProductTx(tx *sqlx.Tx) ProductTx {
	return ProductTx{
		Tx: tx,
	}
}

func (tx ProductTx) DeactivateSiblingProduct(p pw.Product) error {
	_, err := tx.NamedExec(
		pw.StmtDeactivateSiblingProducts,
		p)

	if err != nil {
		return err
	}

	return nil
}

func (tx ProductTx) ActivateProduct(p pw.Product) error {
	_, err := tx.NamedExec(
		pw.StmtActivateProduct,
		p)

	if err != nil {
		return err
	}

	return nil
}

func (tx ProductTx) SetProductOnPaywall(p pw.Product) error {
	_, err := tx.NamedExec(
		pw.StmtPutProductOnPaywall,
		p)

	if err != nil {
		return err
	}

	return nil
}
