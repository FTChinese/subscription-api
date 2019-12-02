package subscription

import (
	"github.com/FTChinese/go-rest/chrono"
	"testing"
	"time"
)

func TestGenerateUpgradeID(t *testing.T) {
	t.Logf("Upgrade id: %s", GenerateUpgradeID())
}

func TestProrationSource_Prorate(t *testing.T) {
	s := ProrationSource{
		PaidAmount: 0.01,
		StartDate:  chrono.DateNow(),
		EndDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
	}

	balance := s.Prorate()

	t.Logf("Balance: %f", balance)
}

func TestNewUpgradePlan(t *testing.T) {
	sources := []ProrationSource{
		{
			PaidAmount: 0.01,
			StartDate:  chrono.DateNow(),
			EndDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
		},
		{
			PaidAmount: 0.01,
			StartDate:  chrono.DateNow(),
			EndDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
		},
		{
			PaidAmount: 0.01,
			StartDate:  chrono.DateNow(),
			EndDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
		},
	}

	up := NewUpgradeIntent(sources)

	t.Logf("Upgrade plan: %+v", up)
}
