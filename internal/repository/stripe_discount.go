package repository

import "github.com/FTChinese/subscription-api/internal/pkg/stripe"

func (repo StripeRepo) UpsertDiscount(d stripe.Discount) error {
	_, err := repo.dbs.Write.NamedExec(stripe.StmtUpsertDiscount, d)

	return err
}

func (repo StripeRepo) RetrieveDiscount(id string) (stripe.Discount, error) {
	var d stripe.Discount
	err := repo.dbs.Read.Get(&d, stripe.StmtRetrieveDiscount, id)

	if err != nil {
		return stripe.Discount{}, err
	}

	return d, nil
}
