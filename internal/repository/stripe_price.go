package repository

import (
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

func (repo StripeRepo) RetrievePrice(id string) (price.StripePrice, error) {
	var p price.StripePrice

	err := repo.dbs.Read.Get(
		&p,
		price.StmtRetrieveStripePrice,
		id)

	if err != nil {
		return price.StripePrice{}, err
	}

	return p, nil
}
