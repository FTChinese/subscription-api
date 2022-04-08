package pw

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
)

// PaywallPrice contains a price's original price and promotion.
// The actual price user paid should be the original price minus
// promotion offer if promotion period is valid.
type PaywallPrice struct {
	price.FtcPrice
	Offers price.DiscountListJSON `json:"offers" db:"discount_list"`
}

// SetOffers refreshes offers saved under a price.
func (p PaywallPrice) SetOffers(o []price.Discount) PaywallPrice {
	p.Offers = o

	return p
}

func (p PaywallPrice) FindValidOffer(id string) (price.Discount, error) {
	if len(p.Offers) == 0 {
		return price.Discount{}, price.ErrDiscountNotFound
	}

	for _, v := range p.Offers {
		if v.ID == id {
			if v.IsValid() {
				return v, nil
			}
		}
	}

	return price.Discount{}, price.ErrDiscountNotFound
}

func (p PaywallPrice) CheckoutItem(offerID null.String) (CartItemFtc, error) {
	// For introductory price, ignore discount.
	if p.IsOneTime() {
		return CartItemFtc{
			Price: p.FtcPrice,
			Offer: price.Discount{},
		}, nil
	}

	if offerID.IsZero() {
		return CartItemFtc{
			Price: p.FtcPrice,
			Offer: price.Discount{},
		}, nil
	}

	offer, err := p.FindValidOffer(offerID.String)
	if err != nil {
		return CartItemFtc{}, err
	}

	return CartItemFtc{
		Price: p.FtcPrice,
		Offer: offer,
	}, nil
}
