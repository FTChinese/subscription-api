package plan

import (
	"gitlab.com/ftchinese/subscription-api/models/util"
	"testing"
)

func TestQuoRem(t *testing.T) {
	t.Logf("%d", 7%10)
	t.Logf("%d", 10%7)

	q, r := util.Division(500.0, 258.0)

	t.Logf("Quotient: %d, remainder: %f", q, r)

	q, r = util.Division(4000.0, 1998.0)

	t.Logf("Quotient: %d, remainder: %f", q, r)
}

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

func TestPlan_CalculateConversion(t *testing.T) {

	type args struct {
		balance float64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Less",
			args: args{
				balance: 451.00,
			},
		},
		{
			name: "Equal",
			args: args{
				balance: 1998.0,
			},
		},
		{
			name: "Exceed",
			args: args{
				balance: 3248.0,
			},
		},
		{
			name: "Near",
			args: args{
				balance: 2000.0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := premiumYearlyPlan.CalculateConversion(tt.args.balance)

			t.Logf("%+v", got)
		})
	}
}

func TestFtcPlans(t *testing.T) {
	t.Logf("FTC live plans: %+v", GetFtcPlans(true))
	t.Logf("FTC test plans: %+v", GetFtcPlans(false))
}
