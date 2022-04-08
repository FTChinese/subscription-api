package pw

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
)

// FtcCartParams contains the item user want to buy.
// Both price and offer only requires id field to be set.
type FtcCartParams struct {
	PriceID    string      `json:"priceId"`
	DiscountID null.String `json:"discountId"`
}

func (s *FtcCartParams) Validate() *render.ValidationError {
	if s.PriceID == "" {
		return &render.ValidationError{
			Message: "Missing priceId field",
			Field:   "priceId",
			Code:    render.CodeMissingField,
		}
	}

	return nil
}

// CartItemFtc contains a price user is trying to purchase and optional discount of this price
// applicable to this user.
type CartItemFtc struct {
	Price price.FtcPrice `json:"price"`
	Offer price.Discount `json:"offer"` // Optional
}

func (i CartItemFtc) PayableAmount() float64 {
	return i.Price.UnitAmount - i.Offer.PriceOff.Float64
}

// PeriodCount selects appropriate period for this purchase.
func (i CartItemFtc) PeriodCount() dt.YearMonthDay {
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
func (i CartItemFtc) Verify(live bool) error {
	// If the discount does not
	if i.Price.ID != i.Offer.PriceID {
		return errors.New("the price and discount does not match")
	}

	if i.Price.LiveMode != live {
		return fmt.Errorf("price defined in %s environment cannot be used in %s environment", ids.GetBoolKey(i.Price.LiveMode), ids.GetBoolKey(live))
	}

	return nil
}
