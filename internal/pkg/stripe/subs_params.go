package stripe

import (
	"fmt"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	stripeSdk "github.com/stripe/stripe-go/v72"
)

// SubsParams is the request body to create a new subscription
// or update an existing one.
// IntroductoryPriceID and CouponID are mutually exclusive.
// When an introductory price exists, a coupon should never be applied.
type SubsParams struct {
	PriceID             string      `json:"priceId"`
	IntroductoryPriceID null.String `json:"introductoryPriceId"`
	CouponID            null.String `json:"coupon"`
	// https://stripe.com/docs/api/subscriptions/create#create_subscription-default_payment_method
	DefaultPaymentMethod null.String `json:"defaultPaymentMethod"`
	// Generated by client. This is optional.
	// It exists to prevent duplicate subscription.
	IdempotencyKey string `json:"idempotency"`
}

// Validate checks if customer and idempotency fields are set.
func (pr SubsParams) Validate() *render.ValidationError {
	if pr.IntroductoryPriceID.Valid || pr.CouponID.Valid {
		return &render.ValidationError{
			Message: "introductory price and coupon cannot be used together",
			Field:   "couponId",
			Code:    render.CodeInvalid,
		}
	}

	return validator.New("priceId").Required().Validate(pr.PriceID)
}

// BuildCartItem constructs a Stripe checkout item
// from cached data.
func (pr SubsParams) BuildCartItem(items []reader.StripePaywallItem) (reader.CartItemStripe, error) {
	index := map[string]int{}
	for i, v := range items {
		index[v.Price.ID] = i
	}

	recurringIndex, ok := index[pr.PriceID]
	if !ok {
		return reader.CartItemStripe{}, fmt.Errorf("stripe price %s not found", pr.PriceID)
	}

	recurringItem := items[recurringIndex]

	var intro price.StripePrice
	var coupon price.StripeCoupon
	if pr.IntroductoryPriceID.Valid {
		introIndex, ok := index[pr.IntroductoryPriceID.String]
		if !ok {
			return reader.CartItemStripe{}, fmt.Errorf("stripe price %s not found", pr.IntroductoryPriceID.String)
		} else {
			intro = items[introIndex].Price
		}
	} else if pr.CouponID.Valid {
		coupon = recurringItem.FindCoupon(pr.CouponID.String)
	}

	return reader.CartItemStripe{
		Recurring:    recurringItem.Price,
		Introductory: intro,
		Coupon:       coupon,
	}, nil
}

func (pr SubsParams) NewSubParams(cusID string, ci reader.CartItemStripe) *stripeSdk.SubscriptionParams {
	params := &stripeSdk.SubscriptionParams{
		Customer:          stripeSdk.String(cusID),
		CancelAtPeriodEnd: stripeSdk.Bool(false),
		Items: []*stripeSdk.SubscriptionItemsParams{
			{
				Price: stripeSdk.String(ci.Recurring.ID),
			},
		},
	}

	// If default payment method is provided, use it;
	// otherwise set payment behavior to incomplete.
	if pr.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripeSdk.String(pr.DefaultPaymentMethod.String)
	} else {
		params.PaymentBehavior = stripeSdk.String("default_incomplete")
	}

	// If there is an introductory offer, add an extra invoice
	// and trial period.
	if !ci.Introductory.IsZero() {
		params.AddInvoiceItems = []*stripeSdk.SubscriptionAddInvoiceItemParams{
			{
				Price:    stripeSdk.String(ci.Introductory.ID),
				Quantity: stripeSdk.Int64(1),
			},
		}

		params.TrialPeriodDays = stripeSdk.Int64(
			ci.Introductory.PeriodCount.TotalDays())
	} else if !ci.Coupon.IsZero() {
		params.Coupon = stripeSdk.String(ci.Coupon.ID)
	}

	// {
	// "status":400,
	// "message":"Idempotent key length is 0 characters long, which is outside accepted lengths. Idempotent Keys must be 1-255 characters long. If you're looking for a decent generator, try using a UUID defined by IETF RFC 4122.",
	// "request_id":"req_O6zILK5QEVpViw",
	// "type":"idempotency_error"
	// }
	if pr.IdempotencyKey != "" {
		params.SetIdempotencyKey(pr.IdempotencyKey)
	}

	// Expand latest_invoice.payment_intent.
	params.AddExpand(KeyLatestInvoicePaymentIntent)

	return params
}

func (pr SubsParams) UpdateSubParams(itemID string, ci reader.CartItemStripe) *stripeSdk.SubscriptionParams {

	params := &stripeSdk.SubscriptionParams{
		CancelAtPeriodEnd: stripeSdk.Bool(false),
		ProrationBehavior: stripeSdk.String(string(stripeSdk.SubscriptionProrationBehaviorCreateProrations)),
		Items: []*stripeSdk.SubscriptionItemsParams{
			{
				// Subscription item to update.
				ID: stripeSdk.String(itemID),
				// The ID of the price object.
				// When changing a subscription item’s price,
				// quantity is set to 1 unless a quantity parameter is provided.
				Price: stripeSdk.String(ci.Recurring.ID),
			},
		},
	}

	// For updating subscription, introductory price should never exist while
	// a coupon is optional.
	if !ci.Coupon.IsZero() {
		params.Coupon = stripeSdk.String(ci.Coupon.ID)
	}

	if pr.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripeSdk.String(pr.DefaultPaymentMethod.String)
	}

	if pr.IdempotencyKey != "" {
		params.SetIdempotencyKey(pr.IdempotencyKey)
	}

	// Expand latest_invoice.payment_intent.
	params.AddExpand(KeyLatestInvoicePaymentIntent)

	return params
}
