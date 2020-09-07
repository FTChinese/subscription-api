package subs

import "github.com/guregu/null"

type Checkout struct {
	PlanID string
	DiscountID null.String
	Amount float64
	CycleCount int
	TrialPeriod int
}
