package model

import "testing"

func TestPlan(t *testing.T) {
	plan := plans["standard_year"]

	t.Log(plan.GetPriceAli())
	t.Log(plan.GetPriceCent())
}

func TestDiscountPlan(t *testing.T) {
	plan, _ := NewPlan(TierStandard, Yearly)

	t.Log(plan)
}
