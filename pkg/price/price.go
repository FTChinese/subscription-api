package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

// Price presents the price of a price. It unified prices coming
// from various source, e.g., FTC in-house or Stripe API.
type Price struct {
	ID string `json:"id" db:"price_id"`
	Edition
	Active      bool        `json:"active" db:"is_active"`
	Archived    bool        `json:"archived" db:"archived"`
	Currency    Currency    `json:"currency" db:"currency"`
	Description null.String `json:"description" db:"description"`
	LiveMode    bool        `json:"liveMode" db:"live_mode"`
	Nickname    null.String `json:"nickname" db:"nickname"`
	ProductID   string      `json:"productId" db:"product_id"`
	Source      Source      `json:"source" db:"source"`
	UnitAmount  float64     `json:"unitAmount" db:"unit_amount"`
	CreatedUTC  chrono.Time `json:"createdUtc" db:"created_utc"`
	CreatedBy   string      `json:"createdBy" db:"created_by"` // Use-facing client should ignore this field.
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
