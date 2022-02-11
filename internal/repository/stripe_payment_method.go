package repository

import "github.com/FTChinese/subscription-api/internal/pkg/stripe"

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
