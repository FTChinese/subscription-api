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

func (repo StripeRepo) RetrieveSetupIntent(id string) (stripe.SetupIntent, error) {
	var si stripe.SetupIntent
	err := repo.dbs.Read.Get(
		&si,
		stripe.RetrieveSetupIntent,
		id)

	if err != nil {
		return stripe.SetupIntent{}, err
	}

	return si, nil
}
