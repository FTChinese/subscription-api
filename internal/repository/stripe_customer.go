package repository

import "github.com/FTChinese/subscription-api/internal/pkg/stripe"

func (repo StripeRepo) InsertCustomer(c stripe.Customer) error {
	_, err := repo.dbs.Write.NamedExec(
		stripe.StmtInsertCustomer,
		c)

	if err != nil {
		return err
	}

	return nil
}

func (repo StripeRepo) UpdateCustomer(c stripe.Customer) error {
	_, err := repo.dbs.Write.NamedExec(stripe.StmtUpdateCustomer, c)

	if err != nil {
		return err
	}

	return nil
}

func (repo StripeRepo) RetrieveCustomer(id string) (stripe.Customer, error) {
	var c stripe.Customer

	err := repo.dbs.Read.Get(&c, stripe.StmtRetrieveCustomer, id)

	if err != nil {
		return stripe.Customer{}, err
	}

	return c, nil
}
