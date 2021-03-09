package price

// Cart contains an item user want to buy and all attributes attached to it like applicable discount, etc..
type Cart struct {
	Price    Price    `json:"price"`
	Discount Discount `json:"discount"`
}

func NewFtcCart(ftcPrice FtcPrice) Cart {
	if ftcPrice.PromotionOffer.IsValid() {
		return Cart{
			Price:    ftcPrice.Price,
			Discount: ftcPrice.PromotionOffer,
		}
	}

	return Cart{
		Price:    ftcPrice.Price,
		Discount: Discount{},
	}
}

// Amount calculates the actual amount user should pay for a plan,
// after taking into account applicable discount, coupon, limited time offer, etc..
func (i Cart) Payable() Charge {
	return Charge{
		Amount:   i.Price.UnitAmount - i.Discount.PriceOff.Float64,
		Currency: string(i.Price.Currency),
	}
}
