package price

import (
	"errors"
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/ids"
)

// FtcPrice contains a price's original price and promotion.
// The actual price user paid should be the original price minus
// promotion offer if promotion period is valid.
type FtcPrice struct {
	Price
	Offers DiscountListJSON `json:"offers" db:"discount_list"`
}

func (f FtcPrice) SetOffers(o []Discount) FtcPrice {
	f.Offers = o
	return f
}

func (f FtcPrice) VerifyOffer(o Discount) error {
	for _, v := range f.Offers {
		if v.ID == o.ID {
			return nil
		}
	}

	return errors.New("the requested offer is not found")
}

// CheckoutItem contains a price user is trying to purchase and optional discount of this price
// applicable to this user.
type CheckoutItem struct {
	Price Price    `json:"price"`
	Offer Discount `json:"offer"` // Optional
}

// Verify checks if the price and offer match after retrieved from
// db separately.
func (i CheckoutItem) Verify(live bool) error {
	// If the discount does not
	if i.Price.ID != i.Offer.PriceID {
		return errors.New("the price and discount does not match")
	}

	if i.Price.LiveMode != live {
		return fmt.Errorf("price defined in %s environment cannot be used in %s environment", ids.GetBoolKey(i.Price.LiveMode), ids.GetBoolKey(live))
	}

	return nil
}
