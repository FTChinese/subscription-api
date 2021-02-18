package product

import "github.com/FTChinese/go-rest/enum"

type PriceSource string

const (
	PriceSourceFTC    = "ftc"
	PriceSourceStripe = "stripe"
)

type Price struct {
	ID string `json:"id"`
	Edition
	Active         bool        `json:"active"`
	Currency       string      `json:"currency"`
	LiveMode       bool        `json:"liveMode"`
	Nickname       string      `json:"nickname"`
	ProductID      string      `json:"productId"`
	Source         PriceSource `json:"source"`
	UnitAmount     float64     `json:"unitAmount"`
	PromotionOffer Discount    `json:"promotionOffer"`
}

func NewPrice(s ExpandedPlanSchema) Price {
	return Price{
		ID:         s.PlanID,
		Edition:    s.Edition,
		Active:     true,
		Currency:   "cny",
		LiveMode:   true,
		Nickname:   "",
		ProductID:  s.ProductID,
		Source:     PriceSourceFTC,
		UnitAmount: s.PlanPrice,
		PromotionOffer: Discount{
			DiscID:        s.DiscID,
			PriceOff:      s.PriceOff,
			Percent:       s.Percent,
			DateTimeRange: s.DateTimeRange,
			Description:   s.Description,
		},
	}
}

func (p Price) DailyCost() DailyCost {
	switch p.Cycle {
	case enum.CycleYear:
		return NewDailyCostOfYear(p.UnitAmount)

	case enum.CycleMonth:
		return NewDailyCostOfMonth(p.UnitAmount)
	}

	return DailyCost{}
}
