package pw

import (
	"github.com/FTChinese/subscription-api/pkg/price"
)

// PriceSchema is used to retrieve a price's price together with
// its discount.
type PriceSchema struct {
	PlanID    string  `db:"plan_id"`
	ProductID string  `db:"product_id"`
	PlanPrice float64 `db:"price"`
	price.Edition
	PlanDesc string `db:"description"`
	price.Discount
}
