package repository

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/internal/pkg"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
)

func (repo StripeRepo) UpsertCustomer(c stripe.Customer) error {
	_, err := repo.dbs.Write.NamedExec(
		stripe.StmtUpsertCustomer,
		c)

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

func (repo StripeRepo) listCusPaymentMethods(cusID string, page gorest.Pagination) ([]stripe.PaymentMethod, error) {
	var paymentMethods = make([]stripe.PaymentMethod, 0)

	err := repo.dbs.Read.Select(
		&paymentMethods,
		stripe.StmtListPaymentMethods,
		cusID,
		page.Limit,
		page.Offset())

	if err != nil {
		return nil, err
	}

	return paymentMethods, nil
}

func (repo StripeRepo) countCusPaymentMethods(cusID string) (int64, error) {
	var total int64
	err := repo.dbs.Read.Get(
		&total,
		cusID)

	if err != nil {
		return -1, err
	}

	return total, nil
}

// ListCusPaymentMethods list a payment methods belongs to a customer.
// NOT a payment method does not belong to any customer
// upon initial creation. It must be attached.
func (repo StripeRepo) ListCusPaymentMethods(cusID string, page gorest.Pagination) (stripe.PagedPaymentMethods, error) {
	defer repo.Logger.Sync()
	sugar := repo.Logger.Sugar()

	countCh := make(chan int64)
	listCh := make(chan stripe.PagedPaymentMethods)

	go func() {
		defer close(countCh)
		n, err := repo.countCusPaymentMethods(cusID)
		if err != nil {
			sugar.Error(err)
		}

		countCh <- n
	}()

	go func() {
		defer close(listCh)
		methods, err := repo.listCusPaymentMethods(cusID, page)

		listCh <- stripe.PagedPaymentMethods{
			PagedList: pkg.PagedList{
				Total:      0,
				Pagination: gorest.Pagination{},
				Err:        err,
			},
			Data: methods,
		}
	}()

	count, listRes := <-countCh, <-listCh

	if listRes.Err != nil {
		return stripe.PagedPaymentMethods{}, listRes.Err
	}

	listRes.PagedList = pkg.PagedList{
		Total:      count,
		Pagination: page,
		Err:        nil,
	}

	return listRes, nil
}
