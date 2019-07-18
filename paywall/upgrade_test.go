package paywall

import (
	"math"
	"testing"
	"time"

	"github.com/FTChinese/go-rest/chrono"
)

var premiumPlan, _ = GetFtcPlans(false).GetPlanByID("premium_year")

func orderID() string {
	id, err := GenerateOrderID()
	if err != nil {
		panic(err)
	}
	return id
}

func buildBalanceSources(count int) []BalanceSource {
	var sources []BalanceSource

	now := time.Now()
	for i := 0; i < count; i++ {
		startTime := now.AddDate(i, 0, 0)
		endTime := startTime.AddDate(i+1, 0, 1)

		s := BalanceSource{
			ID:        orderID(),
			NetPrice:  258.00,
			StartDate: chrono.DateFrom(startTime),
			EndDate:   chrono.DateFrom(endTime),
		}

		sources = append(sources, s)
	}

	return sources
}

func TestBalanceSource_Balance(t *testing.T) {

	tests := []struct {
		name   string
		source BalanceSource
	}{
		{
			name: "Unused Balance Source",
			source: BalanceSource{
				ID:        orderID(),
				NetPrice:  258.00,
				StartDate: chrono.DateNow(),
				EndDate:   chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
			},
		},
		{
			name: "Partly Used Balance Source",
			source: BalanceSource{
				ID:        orderID(),
				NetPrice:  258.00,
				StartDate: chrono.DateFrom(time.Now().AddDate(0, -5, 0)),
				EndDate:   chrono.DateFrom(time.Now().AddDate(0, 7, 1)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Logf("Duration %s to %s", tt.source.StartDate, tt.source.EndDate)
			got := tt.source.Balance()

			t.Logf("Remaining amount: %f", got)
		})
	}
}

func TestUpgrade_SetBalance(t *testing.T) {

	type fields struct {
		Upgrade Upgrade
	}
	type args struct {
		sources []BalanceSource
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Not Enough Balance",
			fields: fields{
				Upgrade: NewUpgrade(premiumPlan),
			},
			args: args{
				sources: buildBalanceSources(2),
			},
		},
		{
			name: "Enough to Cover",
			fields: fields{
				Upgrade: NewUpgrade(premiumPlan),
			},
			args: args{
				sources: buildBalanceSources(8),
			},
		},
		{
			name: "More than 2 cycles",
			fields: fields{
				Upgrade: NewUpgrade(premiumPlan),
			},
			args: args{
				sources: buildBalanceSources(16),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.fields.Upgrade
			got := p.
				SetBalance(tt.args.sources).
				CalculatePayable()

			t.Logf("Upgrade: %+v", got)
		})
	}
}

func TestCalculateCycles(t *testing.T) {
	balance := 2064.00
	price := 1998.00

	cycleCount := 0

	for balance > price {
		cycleCount = cycleCount + 1
		balance = balance - price
	}

	t.Logf("Balance %f, cycle cout %d", balance, cycleCount)

	days := math.Ceil(balance * 365 / price)

	t.Logf("Days: %d", int64(days))
}
