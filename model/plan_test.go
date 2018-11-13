package model

import (
	"testing"

	cache "github.com/patrickmn/go-cache"
)

func TestPlan(t *testing.T) {
	plan := DefaultPlans["standard_year"]

	t.Log(plan.GetPriceString())
	t.Log(plan.GetPriceCent())
}

func TestDiscountPlan(t *testing.T) {
	plan, _ := devEnv.FindPlan(TierStandard, Yearly)

	t.Log(plan)
}

func TestRetrieveSchedule(t *testing.T) {
	sch, err := devEnv.RetrieveSchedule()

	if err != nil {
		t.Error(err)
	}

	t.Log(sch)
}

func TestCache(t *testing.T) {
	c := cache.New(cache.DefaultExpiration, 0)

	c.Set("discountSchedule", DiscountSchedule, cache.NoExpiration)

	sch, found := c.Get("discountSchedule")

	if !found {
		t.Log("Not Found")
		return
	}

	schedule, ok := sch.(Schedule)

	if ok {
		t.Log(schedule)
	}
}
