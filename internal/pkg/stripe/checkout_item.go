package stripe

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/stripe/stripe-go/v72"
)

type CheckoutItem struct {
	Price        Price
	Introductory Price // This is optional.
}

// Validate ensures introductory price is correctly set.
func (ci CheckoutItem) Validate() *render.ValidationError {
	if ci.Introductory.IsZero() {
		return nil
	}

	// You two prices must belong to the same product.
	if ci.Price.ProductID != ci.Introductory.ProductID {
		return &render.ValidationError{
			Message: "Mismatched introductory price",
			Field:   "introductory.product",
			Code:    render.CodeInvalid,
		}
	}

	if ci.Introductory.Kind != price.KindOneTime {
		return &render.ValidationError{
			Message: "Introductory price must be of type one time",
			Field:   "introductory.kind",
			Code:    render.CodeInvalid,
		}
	}

	return nil
}

// NewSubParams build stripe subscription parameters based on the item to check out.
func (ci CheckoutItem) NewSubParams(cusID string, other SubsParams) *stripe.SubscriptionParams {
	params := &stripe.SubscriptionParams{
		Customer:          stripe.String(cusID),
		CancelAtPeriodEnd: stripe.Bool(false),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(ci.Price.ID),
			},
		},
	}

	// If default payment method is provided, use it;
	// otherwise set payment behavior to incomplete.
	if other.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(other.DefaultPaymentMethod.String)
	} else {
		params.PaymentBehavior = stripe.String("default_incomplete")
	}

	// If there is introductory offer, add an extra invoice
	// and trial period.
	if !ci.Introductory.IsZero() {
		params.AddInvoiceItems = []*stripe.SubscriptionAddInvoiceItemParams{
			{
				Price:    stripe.String(ci.Introductory.ID),
				Quantity: stripe.Int64(1),
			},
		}

		params.TrialPeriodDays = stripe.Int64(
			ci.Introductory.PeriodCount.TotalDays())
	}

	// {
	// "status":400,
	// "message":"Idempotent key length is 0 characters long, which is outside accepted lengths. Idempotent Keys must be 1-255 characters long. If you're looking for a decent generator, try using a UUID defined by IETF RFC 4122.",
	// "request_id":"req_O6zILK5QEVpViw",
	// "type":"idempotency_error"
	// }
	if other.IdempotencyKey != "" {
		params.SetIdempotencyKey(other.IdempotencyKey)
	}

	if other.CouponID.Valid {
		params.Coupon = stripe.String(other.CouponID.String)
	}

	// Expand latest_invoice.payment_intent.
	params.AddExpand(KeyLatestInvoicePaymentIntent)

	return params
}

func (ci CheckoutItem) UpdateSubParams(ss *stripe.Subscription, other SubsParams) *stripe.SubscriptionParams {

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
		ProrationBehavior: stripe.String(string(stripe.SubscriptionProrationBehaviorCreateProrations)),
		Items: []*stripe.SubscriptionItemsParams{
			{
				// Subscription item to update.
				ID: stripe.String(ss.Items.Data[0].ID),
				// The ID of the price object.
				// When changing a subscription item’s price,
				// quantity is set to 1 unless a quantity parameter is provided.
				Price: stripe.String(ci.Price.ID),
			},
		},
	}

	if other.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(other.DefaultPaymentMethod.String)
	}

	if other.IdempotencyKey != "" {
		params.SetIdempotencyKey(other.IdempotencyKey)
	}

	if other.CouponID.Valid {
		params.Coupon = stripe.String(other.CouponID.String)
	}

	// Expand latest_invoice.payment_intent.
	params.AddExpand(KeyLatestInvoicePaymentIntent)

	return params
}
