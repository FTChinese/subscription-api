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

// FtcPrice turns data retrieved from db into output format.
// The result FtcPrice's offers retrieve from db
// will be merged with price.FtcOffers.
func (s PriceSchema) FtcPrice() price.FtcPrice {

	var discounts = make([]price.Discount, 0)

	if !s.Discount.IsZero() {
		s.Discount.Kind = price.OfferKindPromotion
		s.Description = null.StringFrom("限时促销")
		discounts = append(discounts, s.Discount)
	}

	// Merge permanent offers.
	offers, ok := price.FtcOffers[s.Edition]
	if ok {
		discounts = append(discounts, offers...)
	}

	return price.FtcPrice{
		Price: price.Price{
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
		Offers:         discounts,
	}
}
