package products

import "github.com/FTChinese/subscription-api/pkg/pw"

func (env Env) ListProducts(live bool) ([]pw.Product, error) {
	var list = make([]pw.Product, 0)
	err := env.dbs.Read.Select(
		&list,
		pw.StmtListProducts,
		live)

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (env Env) CreateProduct(p pw.Product) error {
	_, err := env.dbs.Write.NamedExec(
		pw.StmtCreateProduct,
		p)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) RetrieveProduct(id string, live bool) (pw.Product, error) {
	var p pw.Product
	err := env.dbs.Read.Get(
		&p,
		pw.StmtRetrieveProduct,
		id,
		live)

	if err != nil {
		return pw.Product{}, err
	}

	return p, nil
}

func (env Env) UpdateProduct(p pw.Product) error {
	_, err := env.dbs.Write.NamedExec(
		pw.StmtUpdateProduct,
		p)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) SetProductOnPaywall(p pw.Product) error {
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
