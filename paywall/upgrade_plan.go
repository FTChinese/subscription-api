package paywall

import (
	"github.com/FTChinese/go-rest/chrono"
	"strings"
)

type UpgradePreview struct {
	ID          string      `json:"id"`
	Balance     float64     `json:"balance"` // Accumulated on all BalanceSource.Balance
	SourceIDs   []string    `json:"sources"` // The order ids which still have portion of days unused.
	CreatedAt   chrono.Time `json:"createdAt"`
	ConfirmedAt chrono.Time `json:"confirmedAt"`
	Plan        Plan        `json:"plan"`
	Member      Membership  `json:"-"`
}

func NewUpgradePreview(sources []BalanceSource) UpgradePreview {
	id, _ := GenerateUpgradeID()

	p := UpgradePreview{
		ID:        id,
		CreatedAt: chrono.TimeNow(),
	}

	for _, v := range sources {
		p.Balance = p.Balance + v.Balance()
		p.SourceIDs = append(p.SourceIDs, v.ID)
	}

	return p
}

func (p UpgradePreview) SourceOrderIDs() string {
	return strings.Join(p.SourceIDs, ",")
}
