package stripeclient

import "github.com/stripe/stripe-go/v72"

func (c Client) FetchPaymentIntent(id string) (*stripe.PaymentIntent, error) {
	return c.sc.PaymentIntents.Get(id, nil)
}
