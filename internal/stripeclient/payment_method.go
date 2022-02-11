package stripeclient

import (
	"github.com/stripe/stripe-go/v72"
)

func (c Client) NewPaymentMethod(card *stripe.PaymentMethodCardParams) (*stripe.PaymentMethod, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	pm, err := c.sc.PaymentMethods.New(&stripe.PaymentMethodParams{
		Params:         stripe.Params{},
		Alipay:         nil,
		AUBECSDebit:    nil,
		BACSDebit:      nil,
		Bancontact:     nil,
		BillingDetails: nil,
		Card:           card,
		EPS:            nil,
		FPX:            nil,
		Giropay:        nil,
		Grabpay:        nil,
		Ideal:          nil,
		InteracPresent: nil,
		OXXO:           nil,
		P24:            nil,
		SepaDebit:      nil,
		Sofort:         nil,
		Type:           stripe.String("card"),
		Customer:       nil,
		PaymentMethod:  nil,
	})

	if err != nil {
		sugar.Error(err)
		return nil, err
	}

	return pm, nil
}

func (c Client) FetchPaymentMethod(id string) (*stripe.PaymentMethod, error) {
	return c.sc.PaymentMethods.Get(id, nil)
}
