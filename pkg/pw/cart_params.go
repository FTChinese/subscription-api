package pw

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/guregu/null"
)

// CartParams contains the item user want to buy.
// Both price and offer only requires id field to be set.
type CartParams struct {
	PriceID    string      `json:"priceId"`
	DiscountID null.String `json:"discountId"`
}

func (s *CartParams) Validate() *render.ValidationError {
	if s.PriceID == "" {
		return &render.ValidationError{
			Message: "Missing priceId field",
			Field:   "priceId",
			Code:    render.CodeMissingField,
		}
	}

	return nil
}
