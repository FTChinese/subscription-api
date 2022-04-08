package repository

import "github.com/FTChinese/subscription-api/internal/pkg/stripe"

func (repo StripeRepo) SaveShoppingSession(s stripe.ShoppingSession) error {
	_, err := repo.dbs.Write.NamedExec(stripe.StmtShoppingSession, s)

	if err != nil {
		return err
	}

	return nil
}
