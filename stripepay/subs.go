package stripepay

import (
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/sub"
)

func CreateSubscription(customerID, planID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: stripe.String(planID),
			},
		},
	}

	return sub.New(params)
}
