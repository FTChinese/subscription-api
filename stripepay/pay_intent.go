package stripepay

import (
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/paymentintent"
)

func CreatePaymentIntent(price int64, customerID string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(price),
		Currency: stripe.String(string(stripe.CurrencyCNY)),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Customer: stripe.String(customerID),
	}

	return paymentintent.New(params)
}
