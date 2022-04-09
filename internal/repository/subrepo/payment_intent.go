package subrepo

import (
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/pkg/price"
)

func (env Env) SavePaymentIntent(pi ftcpay.PaymentIntentSchema) error {
	_, err := env.dbs.Write.NamedExec(
		ftcpay.StmtSavePaymentIntent,
		pi)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) RetrieveOrderPrice(orderID string) (price.JSONPrice, error) {
	var p price.JSONPrice

	err := env.dbs.Read.Get(&p, ftcpay.StmtOrderPrice, orderID)

	if err != nil {
		return price.JSONPrice{}, err
	}

	return p, nil
}
