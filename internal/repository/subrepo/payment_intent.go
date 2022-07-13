package subrepo

import (
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
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

func (env Env) RetrievePaymentIntent(orderID string) (ftcpay.PaymentIntent, error) {
	var intent ftcpay.PaymentIntentSchema

	err := env.dbs.Read.Get(
		&intent,
		ftcpay.StmtRetrievePaymentIntent,
		orderID)

	if err != nil {
		return ftcpay.PaymentIntent{}, err
	}

	return ftcpay.PaymentIntent{
		Price: intent.Price.FtcPrice,
		Offer: intent.Offer.Discount,
		Order: ftcpay.Order{
			ID: intent.OrderID,
		},
		Membership: intent.Membership.Membership,
	}, nil
}
