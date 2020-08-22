package stripe

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/validator"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/ephemeralkey"
	"strings"
)

// CreateCustomer sends request to Stripe API to create a customer and returns the customer's id.
func CreateCustomer(email string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
	}

	cus, err := customer.New(params)

	if err != nil {
		return "", err
	}

	return cus.ID, nil
}

// GetDefaultPaymentMethod gets a stripe customer's default payment method.
func GetDefaultPaymentMethod(cusID string) (*stripe.PaymentMethod, error) {
	cus, err := customer.Get(cusID, nil)
	if err != nil {
		return nil, err
	}
	// If InvoiceSettings is nil.
	if cus.InvoiceSettings == nil {
		return nil, nil
	}

	// PaymentInput method might be nil.
	return cus.InvoiceSettings.DefaultPaymentMethod, nil
}

// PaymentInput contains a customer's payment method.
type PaymentInput struct {
	CustomerID    string
	DefaultMethod string `json:"defaultPaymentMethod"`
}

func (p *PaymentInput) Validate() *render.ValidationError {
	p.CustomerID = strings.TrimSpace(p.CustomerID)
	p.DefaultMethod = strings.TrimSpace(p.DefaultMethod)

	ve := validator.New("customerId").Required().Validate(p.CustomerID)
	if ve != nil {
		return ve
	}

	return validator.New("defaultPaymentMethod").Required().Validate(p.DefaultMethod)
}

func SetDefaultPaymentMethod(pm PaymentInput) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm.DefaultMethod),
		},
	}

	return customer.Update(pm.CustomerID, params)
}

func GetEphemeralKey(cusID, version string) ([]byte, error) {
	params := &stripe.EphemeralKeyParams{
		Customer:      stripe.String(cusID),
		StripeVersion: stripe.String(version),
	}

	key, err := ephemeralkey.New(params)
	if err != nil {
		return nil, err
	}

	return key.RawJSON, nil
}
