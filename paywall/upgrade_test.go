package paywall

import (
	"testing"
	"time"

	"github.com/FTChinese/go-rest/chrono"
)

func TestUpgradePlan_CalculatePayable(t *testing.T) {

	plan := GetDefaultPricing()["premium_year"]

	type fields struct {
		Plan       Plan
		Balance    float64
		CycleCount int64
		ExtraDays  int64
		Payable    float64
		OrderIDs   []string
	}
	tests := []struct {
		name   string
		fields fields
		want   UpgradePlan
	}{
		{
			name: "Exactly Cover One Cycle",
			fields: fields{
				Plan:    plan,
				Balance: 1998,
			},
		},
		{
			name: "More Than One Cycle",
			fields: fields{
				Plan:    plan,
				Balance: 2000,
			},
		},
		{
			name: "One Cycle Plus Many Days",
			fields: fields{
				Plan:    plan,
				Balance: 3000,
			},
		},
		{
			name: "Two Cycle Plus Many Days",
			fields: fields{
				Plan:    plan,
				Balance: 4000,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := UpgradePlan{
				Plan:       tt.fields.Plan,
				Balance:    tt.fields.Balance,
				CycleCount: tt.fields.CycleCount,
				ExtraDays:  tt.fields.ExtraDays,
				Payable:    tt.fields.Payable,
				OrderIDs:   tt.fields.OrderIDs,
			}
			//if got := p.CalculatePayable(); !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("UpgradePlan.CalculatePayable() = %v, want %v", got, tt.want)
			//}

			got := p.CalculatePayable()

			t.Logf("Payable: %+v", got)
		})
	}
}

//func TestUnusedOrder_Balance(t *testing.T) {
//	order := UnusedOrder{
//		NetPrice: 258.0,
//		StartDate: chrono.DateFrom(time.Now().Truncate(24*time.Hour)),
//		EndDate: chrono.DateFrom(time.Now().AddDate(0, 1, 0).Truncate(24*time.Hour)),
//	}
//
//	t.Logf("Balance: %f", order.Balance())
//}

func TestUnusedOrder_Balance(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)

	type fields struct {
		ID        string
		NetPrice  float64
		StartDate chrono.Date
		EndDate   chrono.Date
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Unused order",
			fields: fields{
				NetPrice:  258.0,
				StartDate: chrono.DateFrom(today),
				EndDate:   chrono.DateFrom(today.AddDate(0, 1, 0)),
			},
		},
		{
			name: "Partly used order",
			fields: fields{
				NetPrice:  258.0,
				StartDate: chrono.DateFrom(today.AddDate(0, -2, 0)),
				EndDate:   chrono.DateFrom(today.AddDate(0, 10, 0)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := UnusedOrder{
				ID:        tt.fields.ID,
				NetPrice:  tt.fields.NetPrice,
				StartDate: tt.fields.StartDate,
				EndDate:   tt.fields.EndDate,
			}
			got := o.Balance()

			t.Logf("Remaining amount: %f", got)
		})
	}
}
