package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/product"
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/customer"
	"github.com/stripe/stripe-go/v71/ephemeralkey"
	"github.com/stripe/stripe-go/v71/plan"
	"github.com/stripe/stripe-go/v71/sub"
	"go.uber.org/zap"
)

type Client struct {
	live   bool
	logger *zap.Logger
}

// GetPlan retrieves stripe plan details depending on the edition selected.
func (c Client) GetPlan(edition product.Edition) (*stripe.Plan, error) {
	p, err := ftcStripe.PlanStore.FindByEditionV2(edition, c.live)
	if err != nil {
		return nil, err
	}

	return plan.Get(p.PriceID, nil)
}

func (c Client) CreateCustomer(email string) (*stripe.Customer, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	params := &stripe.CustomerParams{
		Email: stripe.String(email),
	}

	cus, err := customer.New(params)

	if err != nil {
		sugar.Error(err)
		return nil, err
	}

	sugar.Infof("New stripe customer: %v", cus)

	return cus, nil
}

// RetrieveCustomer retrieves the details of a stripe customer.
// You can get a customer's default payment method from invoice_settings.default_payment_method.
func (c Client) RetrieveCustomer(cusID string) (*stripe.Customer, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	cus, err := customer.Get(cusID, nil)
	if err != nil {
		sugar.Error(err)
		return nil, err
	}

	sugar.Infof("Retrieve stripe customer: %v", cus)

	return cus, nil
}

// SetDefaultPaymentMethod changes customer's default payment method.
func (c Client) SetDefaultPaymentMethod(pm ftcStripe.PaymentInput) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm.DefaultMethod),
		},
	}

	return customer.Update(pm.CustomerID, params)
}

// CreateEphemeralKey generate a key so that client could restricted customer API directly.
func (c Client) CreateEphemeralKey(cusID, version string) ([]byte, error) {
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

// CreateSubs create a new subscription for a customer.
func (c Client) CreateSubs(cusID string, opts ftcStripe.SubsInput) (*stripe.Subscription, error) {
	plan, err := ftcStripe.PlanStore.FindByEditionV2(opts.Edition, c.live)
	if err != nil {
		return nil, err
	}

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(cusID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: stripe.String(plan.PriceID),
			},
		},
	}

	// {
	// "status":400,
	// "message":"Idempotent key length is 0 characters long, which is outside accepted lengths. Idempotent Keys must be 1-255 characters long. If you're looking for a decent generator, try using a UUID defined by IETF RFC 4122.",
	// "request_id":"req_O6zILK5QEVpViw",
	// "type":"idempotency_error"
	// }
	if opts.IdempotencyKey != "" {
		params.SetIdempotencyKey(opts.IdempotencyKey)
	}

	if opts.CouponID.Valid {
		params.Coupon = stripe.String(opts.DefaultPaymentMethod.String)
	}

	if opts.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(opts.DefaultPaymentMethod.String)
	}

	return sub.New(params)
}

// CancelSubs cancels a subscription at current period end.
func (c Client) CancelSubs(subID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	return sub.Update(subID, params)
}

// UpgradeSubs switch subscription from standard to premium.
func (c Client) UpgradeSubs(subID string, opts ftcStripe.SubsInput) (*stripe.Subscription, error) {
	plan, err := ftcStripe.PlanStore.FindByEditionV2(opts.Edition, c.live)
	if err != nil {
		return nil, err
	}

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: stripe.String(plan.PriceID),
			},
		},
	}

	params.AddExpand("latest_invoice.payment_intent")

	if opts.IdempotencyKey != "" {
		params.IdempotencyKey = stripe.String(opts.IdempotencyKey)
	}

	if opts.CouponID.Valid {
		params.Coupon = stripe.String(opts.CouponID.String)
	}

	if opts.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(opts.DefaultPaymentMethod.String)
	}

	params.SetIdempotencyKey(opts.IdempotencyKey)

	return sub.Update(subID, params)
}
