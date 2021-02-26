package striperepo

import (
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
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
func (c Client) NewSubs(cfg ftcStripe.SubsParams) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer:          stripe.String(cfg.Account.StripeID.String),
		CancelAtPeriodEnd: stripe.Bool(false),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(cfg.Edition.PriceID),
			},
		},
	}

	// Expand latest_invoice.
	params.AddExpand(expandPI)

	// {
	// "status":400,
	// "message":"Idempotent key length is 0 characters long, which is outside accepted lengths. Idempotent Keys must be 1-255 characters long. If you're looking for a decent generator, try using a UUID defined by IETF RFC 4122.",
	// "request_id":"req_O6zILK5QEVpViw",
	// "type":"idempotency_error"
	// }
	if cfg.IdempotencyKey != "" {
		params.SetIdempotencyKey(cfg.IdempotencyKey)
	}

	if cfg.CouponID.Valid {
		params.Coupon = stripe.String(cfg.DefaultPaymentMethod.String)
	}

	if cfg.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(cfg.DefaultPaymentMethod.String)
	}

	return c.sc.Subscriptions.New(params)
}

// UpdateSubs switches subscription billing cycle,
// or upgrading from standard to premium.
func (c Client) UpdateSubs(subID string, cfg ftcStripe.SubsParams) (*stripe.Subscription, error) {
	// Retrieve the subscription first.
	ss, err := c.sc.Subscriptions.Get(subID, nil)
	if err != nil {
		return nil, err
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
		ProrationBehavior: stripe.String(string(stripe.SubscriptionProrationBehaviorCreateProrations)),
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(ss.Items.Data[0].ID),
				Price: stripe.String(cfg.Edition.PriceID),
			},
		},
	}

	// Expand latest_invoice.
	params.AddExpand(expandPI)

	if cfg.IdempotencyKey != "" {
		params.SetIdempotencyKey(cfg.IdempotencyKey)
	}

	if cfg.CouponID.Valid {
		params.Coupon = stripe.String(cfg.DefaultPaymentMethod.String)
	}

	if cfg.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(cfg.DefaultPaymentMethod.String)
	}

	return c.sc.Subscriptions.Update(subID, params)
}

func (c Client) GetSubs(subID string) (*stripe.Subscription, error) {
	p := stripe.Params{}
	p.AddExpand(expandPI)

	return c.sc.Subscriptions.Get(subID, &stripe.SubscriptionParams{
		Params: p,
	})
}

// CancelSubs cancels a subscription at current period end if the passed in parameter `cancel` is true, or reactivate it if false.
func (c Client) CancelSubs(subID string, cancel bool) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(cancel),
	}

	return c.sc.Subscriptions.Update(subID, params)
}
