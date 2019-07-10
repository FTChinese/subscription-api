package billing

import "github.com/FTChinese/go-rest/chrono"

type Upgrade struct {
	ID          string `json:"id"`
	Balance     float64
	Source      []string    `json:"-"`
	CreatedAt   chrono.Time `json:"createdAt"`
	ConfirmedAt chrono.Time `json:"confirmedAt"`
	Member      Member
}
