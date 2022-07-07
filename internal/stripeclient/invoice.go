package stripeclient

import sdk "github.com/stripe/stripe-go/v72"

func (c Client) FetchInvoice(id string) (*sdk.Invoice, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	inv, err := c.sc.Invoices.Get(id, nil)
	if err != nil {
		sugar.Error(err)
		return nil, err
	}

	return inv, nil
}
