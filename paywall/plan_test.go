package paywall

import (
	"testing"
)

func TestPlan_BuildUpgradePlan(t *testing.T) {

	type args struct {
		balance float64
	}
	tests := []struct {
		name string
		plan Plan
		args args
	}{
		{
			name: "Exceed List Price",
			plan: premiumYearlyPlan,
			args: args{
				balance: 3000.00,
			},
		},
		{
			name: "Below List Price",
			plan: premiumYearlyPlan,
			args: args{
				balance: 1000.00,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.plan.WithUpgrade(tt.args.balance)

			t.Logf("Upgrade plan: %+v", got)
		})
	}
}
