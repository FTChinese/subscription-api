package price

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

type Source string

const (
	SourceFTC    = "ftc"
	SourceStripe = "stripe"
)

// Price presents the price of a price. It unified prices coming
// from various source, e.g., FTC in-house or Stripe API.
type Price struct {
	ID string `json:"id"`
	Edition
	Active     bool        `json:"active"`
	Currency   Currency    `json:"currency"`
	LiveMode   bool        `json:"liveMode"`
	Nickname   null.String `json:"nickname"`
	ProductID  string      `json:"productId"`
	Source     Source      `json:"source"`
	UnitAmount float64     `json:"unitAmount"`
}

// DailyCost calculates the daily average price depending on the cycles.
func (p Price) DailyCost() DailyCost {
	switch p.Cycle {
	case enum.CycleYear:
		return NewDailyCostOfYear(p.UnitAmount)

	case enum.CycleMonth:
		return NewDailyCostOfMonth(p.UnitAmount)
	}

	return DailyCost{}
}
