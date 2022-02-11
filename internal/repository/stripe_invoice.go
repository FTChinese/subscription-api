package repository

import "github.com/FTChinese/subscription-api/internal/pkg/stripe"

// UpsertInvoice inserts or updates an invoice.
func (repo StripeRepo) UpsertInvoice(i stripe.Invoice) error {
	_, err := repo.dbs.Write.NamedExec(
		stripe.StmtUpsertInvoice,
		i)
	if err != nil {
		return err
	}

	return nil
}
