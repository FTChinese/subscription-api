package products

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
)

func (env Env) ListProducts(live bool) ([]reader.Product, error) {
	var list = make([]reader.Product, 0)
	err := env.dbs.Read.Select(
		&list,
		reader.StmtListProducts,
		live)

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (env Env) CreateProduct(p reader.Product) error {
	_, err := env.dbs.Write.NamedExec(
		reader.StmtCreateProduct,
		p)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) RetrieveProduct(id string, live bool) (reader.Product, error) {
	var p reader.Product
	err := env.dbs.Read.Get(
		&p,
		reader.StmtRetrieveProduct,
		id,
		live)

	if err != nil {
		return reader.Product{}, err
	}

	return p, nil
}

func (env Env) UpdateProduct(p reader.Product) error {
	_, err := env.dbs.Write.NamedExec(
		reader.StmtUpdateProduct,
		p)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) SetProductOnPaywall(p reader.Product) error {
	tx, err := env.beginProductTx()
	if err != nil {
		return err
	}

	err = tx.DeactivateSiblingProduct(p)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.ActivateProduct(p)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Ignore products in test mode since the paywall table
	// is only designed for live data.
	if p.LiveMode {
		err := tx.SetProductOnPaywallLegacy(p)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	err = tx.SetProductOnPaywall(p)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (env Env) SetProductIntro(p reader.Product) error {
	_, err := env.dbs.Write.NamedExec(
		reader.StmtSetProductIntro,
		p)

	if err != nil {
		return err
	}

	return nil
}
