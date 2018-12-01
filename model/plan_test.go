package model

import (
	"testing"
)

func TestPlan(t *testing.T) {
	plan := DefaultPlans["standard_year"]

	t.Log(plan.GetPriceString())
	t.Log(plan.GetPriceCent())
}
