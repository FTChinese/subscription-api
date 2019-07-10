package billing

import (
	"github.com/FTChinese/go-rest/chrono"
)

// BalanceSource represents user's order
// that is still valid.
type BalanceSource struct {
	ID        string  `json:"id"`
	Amount    float64 `json:"amount"`
	StartDate chrono.Date
	EndDate   chrono.Date
}
