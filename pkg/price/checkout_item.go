package price

import (
	"errors"
	"fmt"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/ids"
)

// CheckoutItem contains a price user is trying to purchase and optional discount of this price
// applicable to this user.
type CheckoutItem struct {
	Price Price    `json:"price"`
	Offer Discount `json:"offer"` // Optional
}

func (i CheckoutItem) PayableAmount() float64 {
	return i.Price.UnitAmount - i.Offer.PriceOff.Float64
}

// PeriodCount selects appropriate period for this purchase.
func (i CheckoutItem) PeriodCount() dt.YearMonthDay {
	if i.Offer.IsZero() {
		return i.Price.PeriodCount.YearMonthDay
	}

	if i.Offer.OverridePeriod.IsZero() {
		return i.Price.PeriodCount.YearMonthDay
	}

	return i.Offer.OverridePeriod.YearMonthDay
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
