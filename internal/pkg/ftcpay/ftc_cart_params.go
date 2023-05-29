package ftcpay

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/guregu/null"
)

// FtcCartParams contains the item user want to buy.
// Both price and offer only requires id field to be set.
type FtcCartParams struct {
	PriceID    string      `json:"priceId"`
	DiscountID null.String `json:"discountId"`
}

func (p *FtcCartParams) Validate() *render.ValidationError {
	if p.PriceID == "" {
		return &render.ValidationError{
			Message: "Missing priceId field",
			Field:   "priceId",
			Code:    render.CodeMissingField,
		}
	}

	return nil
}
