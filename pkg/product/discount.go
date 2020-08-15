package product

import "github.com/guregu/null"

type Discount struct {
	// The id fields started with Disc to avoid conflict when used in ExpandedPlanSchema.
	DiscID   null.String `json:"id" db:"discount_id"`
	PriceOff null.Float  `json:"priceOff" db:"price_off"`
	Percent  null.Int    `json:"percent" db:"percent"`
	Period
}
