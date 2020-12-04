package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/product"
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/client"
	"go.uber.org/zap"
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

// GetPlan retrieves stripe plan details depending on the edition selected.
func (c Client) GetPlan(edition product.Edition) (*stripe.Plan, error) {
	p, err := ftcStripe.PlanStore.FindByEditionV2(edition, c.live)
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
func (c Client) CreateSubs(params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	return c.sc.Subscriptions.New(params)
}

func (c Client) RetrieveSubs(subID string) (*stripe.Subscription, error) {
	return c.sc.Subscriptions.Get(subID, nil)
}

// CancelSubs cancels a subscription at current period end.
func (c Client) CancelSubs(subID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	return c.sc.Subscriptions.Update(subID, params)
}

// UndoCancel set cancel_at_period_end to false before current_period_end past.
func (c Client) UndoCancel(subID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	}

	return c.sc.Subscriptions.Update(subID, params)
}

// UpgradeSubs switch subscription from standard to premium.
func (c Client) UpgradeSubs(subID string, params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	return c.sc.Subscriptions.Update(subID, params)
}
