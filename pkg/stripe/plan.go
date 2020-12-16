package stripe

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/product"
)

var editionKeySuffix = map[bool]string{
	true:  "live",
	false: "test",
}

// Plan contains ftc associates ftc product edition with Stripe
// plan/price id.
// Deprecated
type Plan struct {
	product.Edition
	PriceID string
	Live    bool
}

// Deprecated
type planStore struct {
	plans        []Plan
	indexEdition map[string]int // Index the plans array by `<tier>_<cycle>_<live | test>`
	indexID      map[string]int // Index the plans array by stripe plan/price id.
}

// Deprecated
func newPlanStore() *planStore {
	s := &planStore{
		plans: []Plan{
			{
				Edition: product.NewStdMonthEdition(),
				PriceID: "plan_FXZYLOEbcvj5Tx",
				Live:    true,
			},
			{
				Edition: product.NewStdMonthEdition(),
				PriceID: "plan_FOdgPTznDwHU4i",
				Live:    false,
			},
			{
				Edition: product.NewStdYearEdition(),
				PriceID: "plan_FXZZUEDpToPlZK",
				Live:    true,
			},
			{
				Edition: product.NewStdYearEdition(),
				PriceID: "plan_FOdfeaqzczp6Ag",
				Live:    false,
			},
			{
				Edition: product.NewPremiumEdition(),
				PriceID: "plan_FXZbv1cDTsUKOg",
				Live:    true,
			},
			{
				Edition: product.NewPremiumEdition(),
				PriceID: "plan_FOde0uAr0V4WmT",
				Live:    false,
			},
		},
		indexEdition: make(map[string]int),
		indexID:      make(map[string]int),
	}

	for i, v := range s.plans {
		key := v.NamedKey() + "_" + editionKeySuffix[v.Live]
		s.indexEdition[key] = i
		s.indexID[v.PriceID] = i
	}

	return s
}

// Deprecated
func (s planStore) FindByEdition(e product.Edition, live bool) (Plan, error) {
	i, ok := s.indexEdition[e.NamedKey()+"_"+editionKeySuffix[live]]
	if !ok {
		return Plan{}, fmt.Errorf("stripe plan for %s is not found", e)
	}

	return s.plans[i], nil
}

// FindByID gets Plan by stripe price id.
// Deprecated
func (s planStore) FindByID(priceID string) (Plan, error) {
	i, ok := s.indexID[priceID]
	if !ok {
		return Plan{}, fmt.Errorf("stripe plan with id %s is not found", priceID)
	}

	return s.plans[i], nil
}

// Deprecated
var PlanStore = newPlanStore()
