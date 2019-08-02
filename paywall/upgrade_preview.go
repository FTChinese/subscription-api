package paywall

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/guregu/null"
	"strings"
)

type UpgradePreview struct {
	ID          string      `json:"id"`
	Balance     float64     `json:"balance"` // Accumulated on all BalanceSource.Balance
	SourceIDs   []string    `json:"sources"` // The order ids which still have portion of days unused.
	OrderID     null.String `json:"orderId"`
	CreatedAt   chrono.Time `json:"-"`
	ConfirmedAt chrono.Time `json:"-"`
	Plan        Plan        `json:"plan"`
	Member      Membership  `json:"-"`
}

func NewUpgradePreview(sources []BalanceSource) UpgradePreview {
	id, _ := GenerateUpgradeID()

	up := UpgradePreview{
		ID:        id,
		SourceIDs: []string{}, // This initializes to an empty array, rather than nil so that when marshaling into JSON, it is [], not null.
		CreatedAt: chrono.TimeNow(),
	}

	for _, v := range sources {
		up.Balance = up.Balance + v.Balance()
		up.SourceIDs = append(up.SourceIDs, v.ID)
	}

	// This is hardcoded. Should refactor in the future.
	up.Plan = premiumYearlyPlan.WithUpgrade(up.Balance)
	return up
}

func (up UpgradePreview) SourceOrderIDs() string {
	return strings.Join(up.SourceIDs, ",")
}

func (up UpgradePreview) ReadableBalance() string {
	return fmt.Sprintf("CNY%.2f", up.Balance)
}
