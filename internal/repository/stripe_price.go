package repository

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/price"
)

func (repo StripeRepo) UpsertPrice(p price.StripePrice) error {
	row := stripe.NewPriceRow(p)

	_, err := repo.dbs.Write.NamedExec(
		stripe.StmtUpsertPrice,
		row)

	if err != nil {
		return err
	}

	return nil
}

func (repo StripeRepo) RetrievePrice(id string) (price.StripePrice, error) {
	var row stripe.PriceRow

	err := repo.dbs.Read.Get(
		&row,
		stripe.StmtRetrievePrice,
		id)

	if err != nil {
		return price.StripePrice{}, err
	}

	return row.Data.StripePrice, nil
}
