package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/product"
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/client"
	"go.uber.org/zap"
)

const (
	expandPI = "latest_invoice.payment_intent"
)

type Client struct {
	live   bool
	sc     *client.API
	logger *zap.Logger
}

func NewClient(live bool, logger *zap.Logger) Client {

	key := config.MustLoadStripeAPIKeys().Pick(live)

	return Client{
		live:   live,
		sc:     client.New(key, nil),
		logger: logger,
	}
}

func (c Client) Live() bool {
	return c.live
}

// GetPlan retrieves stripe plan details depending on the edition selected.
// Deprecated
func (c Client) GetPlan(edition product.Edition) (*stripe.Plan, error) {
	p, err := ftcStripe.PlanStore.FindByEdition(edition, c.live)
	if err != nil {
		return nil, err
	}

	return c.sc.Plans.Get(p.PriceID, nil)
}

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

// RetrieveCustomer retrieves the details of a stripe customer.
// You can get a customer's default payment method from invoice_settings.default_payment_method.
func (c Client) RetrieveCustomer(cusID string) (*stripe.Customer, error) {
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

// SetDefaultPaymentMethod changes customer's default payment method.
func (c Client) SetDefaultPaymentMethod(pm ftcStripe.PaymentInput) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm.DefaultMethod),
		},
	}

	return c.sc.Customers.Update(pm.CustomerID, params)
}

func (c Client) CreateSetupIntent(cusID string) (*stripe.SetupIntent, error) {
	params := &stripe.SetupIntentParams{
		Customer: stripe.String(cusID),
		Usage:    stripe.String("on_session"),
	}
	return c.sc.SetupIntents.New(params)
}

// CreateEphemeralKey generate a key so that client could restricted customer API directly.
func (c Client) CreateEphemeralKey(cusID, version string) ([]byte, error) {
	params := &stripe.EphemeralKeyParams{
		Customer:      stripe.String(cusID),
		StripeVersion: stripe.String(version),
	}

	key, err := c.sc.EphemeralKeys.New(params)
	if err != nil {
		return nil, err
	}

	return key.RawJSON, nil
}

// CreateSubs create a new subscription for a customer.
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
func (c Client) CreateSubs(cfg ftcStripe.PaymentConfig) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer:          stripe.String(cfg.Account.StripeID.String),
		CancelAtPeriodEnd: stripe.Bool(false),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: stripe.String(cfg.Plan.PriceID),
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
	if cfg.Params.IdempotencyKey != "" {
		params.SetIdempotencyKey(cfg.Params.IdempotencyKey)
	}

	if cfg.Params.CouponID.Valid {
		params.Coupon = stripe.String(cfg.Params.DefaultPaymentMethod.String)
	}

	if cfg.Params.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(cfg.Params.DefaultPaymentMethod.String)
	}

	return c.sc.Subscriptions.New(params)
}

// UpgradeSubs switch subscription from standard to premium.
func (c Client) UpgradeSubs(subID string, cfg ftcStripe.PaymentConfig) (*stripe.Subscription, error) {
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
				Price: stripe.String(cfg.Plan.PriceID),
			},
		},
	}

	// Expand latest_invoice.
	params.AddExpand(expandPI)

	if cfg.Params.IdempotencyKey != "" {
		params.SetIdempotencyKey(cfg.Params.IdempotencyKey)
	}

	if cfg.Params.CouponID.Valid {
		params.Coupon = stripe.String(cfg.Params.DefaultPaymentMethod.String)
	}

	if cfg.Params.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(cfg.Params.DefaultPaymentMethod.String)
	}

	return c.sc.Subscriptions.Update(subID, params)
}

func (c Client) RetrieveSubs(subID string) (*stripe.Subscription, error) {
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

func (c Client) ListPrices() []*stripe.Price {
	return c.sc.Prices.List(&stripe.PriceListParams{
		Active: stripe.Bool(true),
	}).PriceList().Data
}
