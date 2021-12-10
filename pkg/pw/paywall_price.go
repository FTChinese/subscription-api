package pw

import (
	"errors"
	"github.com/FTChinese/subscription-api/pkg/price"
)

// PaywallPrice contains a price's original price and promotion.
// The actual price user paid should be the original price minus
// promotion offer if promotion period is valid.
type PaywallPrice struct {
	price.Price
	Offers price.DiscountListJSON `json:"offers" db:"discount_list"`
}

func (p PaywallPrice) SetOffers(o []price.Discount) PaywallPrice {
	p.Offers = o

	return p
}

func (p PaywallPrice) VerifyOffer(o price.Discount) error {
	for _, v := range p.Offers {
		if v.ID == o.ID {
			return nil
		}
	}

	return errors.New("the requested offer is not found")
}
