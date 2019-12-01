package plan

import "github.com/FTChinese/go-rest/chrono"

type UpgradeScheme struct {
	ID        string            `json:"id" db:"upgrade_id"`
	Balance   float64           `json:"balance" db:"balance"`
	CreatedAt chrono.Time       `json:"createdAt" db:"created_at"`
	Source    []ProrationSource `json:"-"`
}
