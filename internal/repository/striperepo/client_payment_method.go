package striperepo

import (
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
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

func (c Client) AttachPaymentMethod(cusID string, pmID string) (*stripe.SetupIntent, error) {
	si, err := c.sc.SetupIntents.New(&stripe.SetupIntentParams{
		Params:               stripe.Params{},
		Confirm:              stripe.Bool(true),
		Customer:             stripe.String(cusID),
		Description:          nil,
		MandateData:          nil,
		OnBehalfOf:           nil,
		PaymentMethod:        stripe.String(pmID),
		PaymentMethodOptions: nil,
		PaymentMethodTypes:   nil,
		ReturnURL:            nil,
		SingleUse:            nil,
		Usage:                stripe.String("off_session"),
	})

	if err != nil {
		return nil, err
	}

	return si, nil
}

// SetDefaultPaymentMethod changes customer's default payment method.
func (c Client) SetDefaultPaymentMethod(pm ftcStripe.PaymentInput) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm.DefaultMethod),
		},
	}

	return c.sc.Customers.Update(pm.CustomerID, params)
}
