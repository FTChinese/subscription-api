package repository

import "github.com/FTChinese/subscription-api/internal/pkg/stripe"

func (repo StripeRepo) UpsertSetupIntent(si stripe.SetupIntent) error {
	_, err := repo.dbs.Write.NamedExec(
		stripe.StmtUpsertSetupIntent,
		si)

	if err != nil {
		return err
	}

	return nil
}
