package addons

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
)

// InvoicesCarriedOver flags a user's consumed
// but not ended invoices as being carried over
// into a new invoice.
// This happens in such a situation:
// A standard user's  addons were already transferred
// to membership, and then upgrade to premium.
// Obviously the addon invoices are not yet consumed completely.
// The remaining times of transferred to a new addon
// invoice. Add a timestamp to those invoices indicating
// that then are again carried over.
// Not much usefulness.
func (env Env) InvoicesCarriedOver(userID ids.UserIDs) error {
	_, err := env.dbs.Write.Exec(
		invoice.StmtCarriedOver,
		userID.BuildFindInSet(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (env Env) countInvoices(ids ids.UserIDs) (int64, error) {
	var count int64
	err := env.dbs.Read.Get(
		&count,
		invoice.StmtCountInvoices,
		ids.BuildFindInSet())

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (env Env) listInvoices(ids ids.UserIDs, p gorest.Pagination) ([]invoice.Invoice, error) {
	var inv = make([]invoice.Invoice, 0)
	err := env.dbs.Read.Select(
		&inv,
		invoice.StmtListInvoices,
		ids.BuildFindInSet(),
		p.Limit,
		p.Offset(),
	)
	if err != nil {
		return nil, err
	}

	return inv, nil
}

func (env Env) ListInvoices(ids ids.UserIDs, p gorest.Pagination) (invoice.List, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	countCh := make(chan int64)
	listCh := make(chan invoice.List)

	go func() {
		defer close(countCh)
		n, err := env.countInvoices(ids)
		if err != nil {
			sugar.Error(err)
		}

		countCh <- n
	}()

	go func() {
		defer close(listCh)
		invs, err := env.listInvoices(ids, p)
		if err != nil {
			sugar.Error(err)
		}
		listCh <- invoice.List{
			Total:      0,
			Pagination: gorest.Pagination{},
			Data:       invs,
			Err:        err,
		}
	}()

	count, listResult := <-countCh, <-listCh

	if listResult.Err != nil {
		return invoice.List{}, listResult.Err
	}

	return invoice.List{
		Total:      count,
		Pagination: p,
		Data:       listResult.Data,
	}, nil
}

func (env Env) LoadInvoice(invID string) (invoice.Invoice, error) {
	var inv invoice.Invoice
	err := env.dbs.Read.Get(
		&inv,
		invoice.StmtSelectInvoice,
		invID)
	if err != nil {
		return invoice.Invoice{}, err
	}

	return inv, nil
}
