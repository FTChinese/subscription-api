package repository

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/internal/pkg"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
)

// UpsertPaymentMethod inserts/update a payment method.
// This usually happens when a default payment method is set
// on a customer or on a subscription.
func (repo StripeRepo) UpsertPaymentMethod(pm stripe.PaymentMethod) error {
	_, err := repo.dbs.Write.NamedExec(
		stripe.StmtInsertPaymentMethod,
		pm)

	if err != nil {
		return err
	}

	return nil
}

func (repo StripeRepo) RetrievePaymentMethod(id string) (stripe.PaymentMethod, error) {
	var pm stripe.PaymentMethod
	err := repo.dbs.Read.Get(
		&pm,
		stripe.StmtRetrievePaymentMethod,
		id)

	if err != nil {
		return stripe.PaymentMethod{}, err
	}

	return pm, nil
}

func (repo StripeRepo) listPaymentMethods(cusID string, page gorest.Pagination) ([]stripe.PaymentMethod, error) {
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

func (repo StripeRepo) countPaymentMethods(cusID string) (int64, error) {
	var total int64
	err := repo.dbs.Read.Get(
		&total,
		cusID)

	if err != nil {
		return -1, err
	}

	return total, nil
}

func (repo StripeRepo) ListPaymentMethods(cusID string, page gorest.Pagination) (stripe.PagedPaymentMethods, error) {
	defer repo.Logger.Sync()
	sugar := repo.Logger.Sugar()

	countCh := make(chan int64)
	listCh := make(chan stripe.PagedPaymentMethods)

	go func() {
		defer close(countCh)
		n, err := repo.countPaymentMethods(cusID)
		if err != nil {
			sugar.Error(err)
		}

		countCh <- n
	}()

	go func() {
		defer close(listCh)
		methods, err := repo.listPaymentMethods(cusID, page)

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
