package billing

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
)

// Plan is a pricing plan.
// The list price is the price that buyers pay for your product or service without any discounts.
// The net price of a product or service is the actual price that customers pay for the product or service.
type Plan struct {
	ID         string      `json:"id"`
	Active     bool        `json:"active"`
	Amount     float64     `json:"amount"`
	Caption    string      `json:"caption"`
	Cycle      enum.Cycle  `json:"cycle"`
	CycleCount int64       `json:"intervalCount"`
	Currency   string      `json:"currency"`
	ProductID  string      `json:"productId"`
	Tier       enum.Tier   `json:"tier"` // This is product.
	TrialDays  int64       `json:"trialDays"`
	CreatedAt  chrono.Time `json:"createdAt"`
}
