package stripeclient

import (
	ftcStripe "github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/stripe/stripe-go/v72"
)

// NewSubs create a new subscription for a customer.
// Error returned in case you missed Customer field:
// {
//    "code": "parameter_missing",
//    "doc_url": "https://stripe.com/docs/error-codes/parameter-missing",
//    "status": 400,
//    "message": "Missing required param: customer.",
//    "param": "customer",
//    "request_id": "req_BthzW5QZzNDTwN",
//    "type": "invalid_request_error"
// }
func (c Client) NewSubs(params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	return c.sc.Subscriptions.New(params)
}

// UpdateSubs switches subscription billing cycle,
// or upgrading from standard to premium.
func (c Client) UpdateSubs(subID string, params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	return c.sc.Subscriptions.Update(subID, params)
}

func (c Client) GetSubs(subID string, expand bool) (*stripe.Subscription, error) {
	var params *stripe.SubscriptionParams
	if expand {
		p := stripe.Params{}
		p.AddExpand(ftcStripe.KeyLatestInvoicePaymentIntent)
	}

	return c.sc.Subscriptions.Get(subID, params)
}

// CancelSubs cancels a subscription at current period end if the passed in parameter `cancel` is true, or reactivate it if false.
func (c Client) CancelSubs(subID string, cancel bool) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(cancel),
	}

	return c.sc.Subscriptions.Update(subID, params)
}
