package repository

import "github.com/FTChinese/subscription-api/internal/pkg/stripe"

func (repo StripeRepo) SaveWebhookError(whe stripe.WebhookError) error {
	_, err := repo.dbs.Write.NamedExec(stripe.StmtInsertWebhookError, whe)

	if err != nil {
		return err
	}

	return nil
}
