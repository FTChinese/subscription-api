package reader

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
)

// CartItemStripe contains the items to be purchased using
// stripe.
type CartItemStripe struct {
	Recurring    price.StripePrice
	Introductory price.StripePrice // This is optional.
}

// AnyFromStripe checks if there's any price coming from API
// so that we could save it.
func (ci CartItemStripe) AnyFromStripe() bool {
	return ci.Recurring.IsFromStripe || ci.Introductory.IsFromStripe
}

// Validate ensures introductory price is correctly set.
func (ci CartItemStripe) Validate() *render.ValidationError {
	if ci.Introductory.IsZero() {
		return nil
	}

	// The two prices must belong to the same product.
	if ci.Recurring.ProductID != ci.Introductory.ProductID {
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
