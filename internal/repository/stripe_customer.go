package repository

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg"
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
		stripe.StmtCountPaymentMethods,
		cusID)

	if err != nil {
		return -1, err
	}

	return total, nil
}

// ListCusPaymentMethods list a payment methods belongs to a customer.
// NOT a payment method does not belong to any customer
// upon initial creation. It must be attached.
func (repo StripeRepo) ListCusPaymentMethods(cusID string, page gorest.Pagination) (pkg.PagedData[stripe.PaymentMethod], error) {
	defer repo.Logger.Sync()
	sugar := repo.Logger.Sugar()

	countCh := make(chan int64)
	listCh := make(chan pkg.AsyncResult[[]stripe.PaymentMethod])

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

		listCh <- pkg.AsyncResult[[]stripe.PaymentMethod]{
			Err:   err,
			Value: methods,
		}
	}()

	count, listRes := <-countCh, <-listCh

	if listRes.Err != nil {
		return pkg.PagedData[stripe.PaymentMethod]{}, listRes.Err
	}

	return pkg.PagedData[stripe.PaymentMethod]{
		Total:      count,
		Pagination: page,
		Data:       listRes.Value,
	}, nil
}
