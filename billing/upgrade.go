package billing

import "github.com/FTChinese/go-rest/chrono"

type Upgrade struct {
	ID          string   `json:"id"`
	Source      []string `json:"-"`
	Balance     float64
	CreatedAt   chrono.Time `json:"createdAt"`
	ConfirmedAt chrono.Time `json:"confirmedAt"`
}
