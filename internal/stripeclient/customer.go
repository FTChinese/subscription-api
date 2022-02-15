package stripeclient

import (
	sdk "github.com/stripe/stripe-go/v72"
)

func (c Client) CreateCustomer(email string) (*sdk.Customer, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	params := &sdk.CustomerParams{
		Email: sdk.String(email),
	}

	cus, err := c.sc.Customers.New(params)

	if err != nil {
		sugar.Error(err)
		return nil, err
	}

	sugar.Infof("New stripe customer: %v", cus)

	return cus, nil
}

// FetchCustomer retrieves the details of a stripe customer.
// You can get a customer's default payment method from invoice_settings.default_payment_method.
func (c Client) FetchCustomer(cusID string) (*sdk.Customer, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	cus, err := c.sc.Customers.Get(cusID, nil)
	if err != nil {
		sugar.Error(err)
		return nil, err
	}

	sugar.Infof("Retrieve stripe customer: %v", cus)

	return cus, nil
}

// SetCusDefaultPaymentMethod changes customer's default payment method.
func (c Client) SetCusDefaultPaymentMethod(cusID string, pmID string) (*sdk.Customer, error) {
	params := &sdk.CustomerParams{
		InvoiceSettings: &sdk.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: sdk.String(pmID),
		},
	}

	return c.sc.Customers.Update(cusID, params)
}
