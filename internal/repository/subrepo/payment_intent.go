package subrepo

import (
	"github.com/FTChinese/subscription-api/internal/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/price"
)

func (env Env) SavePaymentIntent(pi subs.PaymentIntentSchema) error {
	_, err := env.dbs.Write.NamedExec(
		subs.StmtSavePaymentIntent,
		pi)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) RetrieveOrderPrice(orderID string) (price.JSONPrice, error) {
	var p price.JSONPrice

	err := env.dbs.Read.Get(&p, subs.StmtOrderPrice, orderID)

	if err != nil {
		return price.JSONPrice{}, err
	}

	return p, nil
}
