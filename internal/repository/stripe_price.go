package repository

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/price"
)

func (repo StripeRepo) UpsertPrice(p price.StripePrice) error {

	_, err := repo.dbs.Write.NamedExec(
		price.StmtUpsertStripePrice,
		p)

	if err != nil {
		return err
	}

	return nil
}

func (repo StripeRepo) RetrievePrice(id string, live bool) (price.StripePrice, error) {
	var p price.StripePrice

	err := repo.dbs.Read.Get(
		&p,
		price.StmtRetrieveStripePrice,
		id,
		live)

	if err != nil {
		return price.StripePrice{}, err
	}

	return p, nil
}

// ActivatePrice adds an entry into product_active_price table.
func (repo StripeRepo) ActivatePrice(p price.StripePrice) error {
	tx, err := txrepo.BeginPriceTx(repo.dbs.Write)
	if err != nil {
		return err
	}

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

func (repo StripeRepo) DeactivePrice(p price.StripePrice) error {
	tx, err := txrepo.BeginPriceTx(repo.dbs.Delete)
	if err != nil {
		return err
	}

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
