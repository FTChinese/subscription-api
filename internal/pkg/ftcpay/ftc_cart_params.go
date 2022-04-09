package ftcpay

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/reader"
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

func (p FtcCartParams) BuildCartItem(products []reader.PaywallProduct) (reader.CartItemFtc, error) {
	pwPrice, err := reader.FindPaywallPrice(products, p.PriceID)

	if err != nil {
		return reader.CartItemFtc{}, err
	}

	return pwPrice.BuildCartItem(p.DiscountID)
}
