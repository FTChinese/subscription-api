package striperepo

import "github.com/FTChinese/subscription-api/pkg/stripe"

func (env Env) UpsertInvoice(i stripe.Invoice) error {
	_, err := env.db.NamedExec(stripe.StmtUpsertInvoice, i)
	if err != nil {
		return err
	}

	return nil
}
