package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/guregu/null"
)

// Plan describes a subscription plan.
// Deprecated. Use Price.
type Plan struct {
	ID string `json:"id" db:"id"`
	Edition
	Active        bool        `json:"active" db:"active"`
	Archived      bool        `json:"archived" db:"archived"` // Once archived, it should never be touched.
	Currency      Currency    `json:"currency" db:"currency"`
	Description   null.String `json:"description" db:"description"`
	LiveMode      bool        `json:"liveMode" db:"live_mode"`
	Nickname      null.String `json:"nickname" db:"nickname"`
	ProductID     string      `json:"productId" db:"product_id"`
	StripePriceID string      `json:"stripePriceId" db:"stripe_price_id"`
	UnitAmount    float64     `json:"unitAmount" db:"unit_amount"`
	CreatedUTC    chrono.Time `json:"createdUtc" db:"created_utc"`
}

// NewPlan creates a new Plan.
// Deprecated.
func NewPlan(p Price) Plan {
	return Plan{
		ID:            p.ID,
		Edition:       p.Edition,
		Active:        p.Active,
		Archived:      p.Archived,
		Currency:      p.Currency,
		Description:   p.Title,
		LiveMode:      p.LiveMode,
		Nickname:      p.Nickname,
		ProductID:     p.ProductID,
		StripePriceID: p.StripePriceID,
		UnitAmount:    p.UnitAmount,
		CreatedUTC:    p.CreatedUTC,
	}
}
