package stripe

import (
	"database/sql"
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/plan"
)

var editionKeySuffix = map[bool]string{
	true:  "live",
	false: "test",
}

// Plan contains ftc associates ftc product edition with Stripe
// plan/price id.
type Plan struct {
	product.Edition
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

func (s planStore) FindByEdition(key string, live bool) (Plan, error) {
	i, ok := s.indexEdition[key+"_"+editionKeySuffix[live]]
	if !ok {
		return Plan{}, fmt.Errorf("stripe plan for %s is not found", key)
	}

	return s.plans[i], nil
}

func (s planStore) FindByEditionV2(e product.Edition, live bool) (Plan, error) {
	i, ok := s.indexEdition[e.NamedKey()+"_"+editionKeySuffix[live]]
	if !ok {
		return Plan{}, fmt.Errorf("stripe plan for %s is not found", e)
	}

	return s.plans[i], nil
}

func (s planStore) FindByID(planID string) (Plan, error) {
	i, ok := s.indexID[planID]
	if !ok {
		return Plan{}, fmt.Errorf("stripe plan with id %s is not found", planID)
	}

	return s.plans[i], nil
}

var PlanStore = newPlanStore()

// FetchPlan gets stripe plan from API.
// The key is one of standard_month, standard_year, premium_year.
func FetchPlan(key string, live bool) (*stripe.Plan, error) {
	p, err := PlanStore.FindByEdition(key, live)
	if err != nil {
		return nil, sql.ErrNoRows
	}

	return plan.Get(p.PriceID, nil)
}
