package stripeclient

import (
	"github.com/stripe/stripe-go/v72"
)

func (c Client) FetchPaymentMethod(id string) (*stripe.PaymentMethod, error) {
	return c.sc.PaymentMethods.Get(id, nil)
}
