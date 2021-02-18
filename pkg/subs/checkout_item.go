package subs

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
)

// CheckoutItem contains an item user want to buy and all attributes attached to it like applicable discount, coupon, etc..
type CheckoutItem struct {
	Price    price.Price    `json:"price"`
	Discount price.Discount `json:"discount"`
}

func NewCheckoutItem(pp pw.ProductPrice) CheckoutItem {
	if pp.PromotionOffer.IsValid() {
		return CheckoutItem{
			Price:    pp.Original,
			Discount: pp.PromotionOffer,
		}
	}

	return CheckoutItem{
		Price:    pp.Original,
		Discount: price.Discount{},
	}
}

// Amount calculates the actual amount user should pay for a plan,
// after taking into account applicable discount, coupon, limited time offer, etc..
func (i CheckoutItem) Payable() price.Charge {
	return price.Charge{
		Amount:   i.Price.UnitAmount - i.Discount.PriceOff.Float64,
		Currency: "cny",
	}
}
