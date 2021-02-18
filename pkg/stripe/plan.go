package stripe

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/price"
)

var editionKeySuffix = map[bool]string{
	true:  "live",
	false: "test",
}

// Plan contains ftc associates ftc price edition with Stripe
// plan/price id.
type Plan struct {
	price.Edition
	PriceID string
	Live    bool
}

type planStore struct {
	plans        []Plan
	indexEdition map[string]int // Index the plans array by `<tier>_<cycle>_<live | test>`
	indexID      map[string]int // Index the plans array by stripe plan/price id.
}

func newPlanStore() *planStore {
	s := &planStore{
		plans: []Plan{
			{
				Edition: price.NewStdMonthEdition(),
				PriceID: "price_1IM2Z4BzTK0hABgJ9Sh0u35h",
				Live:    true,
			},
			{
				Edition: price.NewStdMonthEdition(),
				PriceID: "price_1IM2mgBzTK0hABgJVH8o9Sjm",
				Live:    false,
			},
			{
				Edition: price.NewStdYearEdition(),
				PriceID: "price_1IM2aNBzTK0hABgJeJVIx3kL",
				Live:    true,
			},
			{
				Edition: price.NewStdYearEdition(),
				PriceID: "price_1IM2nFBzTK0hABgJiIDeDIox",
				Live:    false,
			},
			{
				Edition: price.NewPremiumEdition(),
				PriceID: "plan_FXZbv1cDTsUKOg",
				Live:    true,
			},
			{
				Edition: price.NewPremiumEdition(),
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

func (s planStore) FindByEdition(e price.Edition, live bool) (Plan, error) {
	i, ok := s.indexEdition[e.NamedKey()+"_"+editionKeySuffix[live]]
	if !ok {
		return Plan{}, fmt.Errorf("stripe plan for %s is not found", e)
	}

	return s.plans[i], nil
}

// FindByID gets Plan by stripe price id.
func (s planStore) FindByID(priceID string) (Plan, error) {
	i, ok := s.indexID[priceID]
	if !ok {
		return Plan{}, fmt.Errorf("stripe plan with id %s is not found", priceID)
	}

	return s.plans[i], nil
}

var PlanStore = newPlanStore()
