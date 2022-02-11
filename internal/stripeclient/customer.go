package stripeclient

import (
	ftcStripe "github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/stripe/stripe-go/v72"
)

func (c Client) CreateCustomer(email string) (*stripe.Customer, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	params := &stripe.CustomerParams{
		Email: stripe.String(email),
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
func (c Client) FetchCustomer(cusID string) (*stripe.Customer, error) {
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
func (c Client) SetCusDefaultPaymentMethod(pm ftcStripe.DefaultPaymentMethodParams) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm.DefaultMethod),
		},
	}

	return c.sc.Customers.Update(pm.CustomerID, params)
}
