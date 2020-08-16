package product

import (
	"github.com/guregu/null"
	"time"
)

type Discount struct {
	// The id fields started with Disc to avoid conflict when used in ExpandedPlanSchema.
	DiscID   null.String `json:"id" db:"discount_id"`
	PriceOff null.Float  `json:"priceOff" db:"price_off"`
	Percent  null.Int    `json:"percent" db:"percent"`
	Period
}

func (d Discount) IsValid() bool {
	if d.PriceOff.IsZero() || d.PriceOff.Float64 <= 0 {
		return false
	}

	now := time.Now()

	if now.Before(d.StartUTC.Time) || now.After(d.EndUTC.Time) {
		return false
	}

	return true
}
