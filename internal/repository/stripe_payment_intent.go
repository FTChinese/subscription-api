package repository

import "github.com/FTChinese/subscription-api/internal/pkg/stripe"

func (repo StripeRepo) UpsertPaymentIntent(pi stripe.PaymentIntent) error {
	_, err := repo.dbs.Write.NamedExec(
		stripe.StmtUpsertPaymentIntent,
		pi)

	if err != nil {
		return err
	}

	return nil
}
