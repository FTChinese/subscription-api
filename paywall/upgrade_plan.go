package paywall

import (
	"github.com/FTChinese/go-rest/chrono"
	"strings"
)

type UpgradePreview struct {
	ID          string      `json:"id"`
	Balance     float64     `json:"balance"` // Accumulated on all BalanceSource.Balance
	SourceIDs   []string    `json:"sources"` // The order ids which still have portion of days unused.
	CreatedAt   chrono.Time `json:"-"`
	ConfirmedAt chrono.Time `json:"-"`
	Plan        Plan        `json:"plan"`
	Member      Membership  `json:"-"`
}

func NewUpgradePreview(sources []BalanceSource) UpgradePreview {
	id, _ := GenerateUpgradeID()

	up := UpgradePreview{
		ID:        id,
		CreatedAt: chrono.TimeNow(),
	}

	for _, v := range sources {
		up.Balance = up.Balance + v.Balance()
		up.SourceIDs = append(up.SourceIDs, v.ID)
	}

	return up
}

func (up UpgradePreview) SourceOrderIDs() string {
	return strings.Join(up.SourceIDs, ",")
}
