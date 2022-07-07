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

func (repo StripeRepo) RetrieveInvoice(id string) (stripe.Invoice, error) {
	var inv stripe.Invoice
	err := repo.dbs.Read.Get(&inv, stripe.StmtRetrieveInvoice, id)
	if err != nil {
		return stripe.Invoice{}, err
	}

	return inv, nil
}
