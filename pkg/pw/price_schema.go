package pw

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
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

// NewFtcPrice turns the raw data from DB into FtcPrice.
func NewFtcPrice(s PriceSchema) price.FtcPrice {
	return price.FtcPrice{
		Original: price.Price{
			ID:         s.PlanID,
			Edition:    s.Edition,
			Active:     true,
			Currency:   price.CurrencyCNY,
			LiveMode:   true,
			Nickname:   null.String{},
			ProductID:  s.ProductID,
			Source:     price.SourceFTC,
			UnitAmount: s.PlanPrice,
		},
		PromotionOffer: s.Discount,
	}
}
