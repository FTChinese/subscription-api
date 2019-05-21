package paywall

import (
	"testing"

	"github.com/Pallinder/go-randomdata"
)

func TestUpgradePlan(t *testing.T) {

	var orders = make([]Proration, 0)
	loop := randomdata.Number(3)
	for i := 0; i < loop; i++ {
		orderID, _ := GenerateOrderID()
		orders = append(orders, Proration{
			OrderID: orderID,
			Balance: randomdata.Decimal(1998),
		})
	}

	plan := GetDefaultPricing()["premium_year"]

	up := NewUpgradePlan(plan).
		SetProration(orders).
		CalculatePayable()

	t.Logf("Upgrade plan %+v", up)
}

//func TestUpgradePlan_CalculatePayable(t *testing.T) {
//	p := UpgradePlan{
//		Balance: 1998,
//	}
//
//	p.Tier = enum.TierPremium
//	p.Cycle = enum.CycleYear
//	p.ListPrice = 1998
//	p.NetPrice = 1998
//
//	p = p.CalculatePayable()
//
//	t.Logf("Payable: %+v", p)
//}

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
