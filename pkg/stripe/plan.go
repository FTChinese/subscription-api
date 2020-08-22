package stripe

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/plan"
)

var stripeKeySuffix = map[bool]string{
	true:  "live",
	false: "test",
}

type PlanConfig struct {
	product.Edition
	PlanID string
	Live   bool
}

// GetPlan fetches stripe plan from stripe API.
func (c PlanConfig) GetPlan() (*stripe.Plan, error) {
	return plan.Get(c.PlanID, nil)
}

type stripeStore struct {
	plans        []PlanConfig
	indexEdition map[string]int
	indexID      map[string]int
}

func newStripeStoreSchema() *stripeStore {
	s := &stripeStore{
		plans: []PlanConfig{
			{
				Edition: product.NewStdMonthEdition(),
				PlanID:  "plan_FXZYLOEbcvj5Tx",
				Live:    true,
			},
			{
				Edition: product.NewStdMonthEdition(),
				PlanID:  "plan_FOdgPTznDwHU4i",
				Live:    false,
			},
			{
				Edition: product.NewStdYearEdition(),
				PlanID:  "plan_FXZZUEDpToPlZK",
				Live:    true,
			},
			{
				Edition: product.NewStdYearEdition(),
				PlanID:  "plan_FOdfeaqzczp6Ag",
				Live:    false,
			},
			{
				Edition: product.NewPremiumEdition(),
				PlanID:  "plan_FXZbv1cDTsUKOg",
				Live:    true,
			},
			{
				Edition: product.NewPremiumEdition(),
				PlanID:  "plan_FOde0uAr0V4WmT",
				Live:    false,
			},
		},
		indexEdition: make(map[string]int),
		indexID:      make(map[string]int),
	}

	for i, v := range s.plans {
		key := v.NamedKey() + "_" + stripeKeySuffix[v.Live]
		s.indexEdition[key] = i
		s.indexID[v.PlanID] = i
	}

	return s
}

func (s *stripeStore) add(p PlanConfig) *stripeStore {
	s.plans = append(s.plans, p)

	pos := len(s.plans) - 1
	key := p.NamedKey() + "_" + stripeKeySuffix[p.Live]

	// Uses <tier>_<cycle>_<live|test> as key.
	// Example: `standard_year_live`
	s.indexEdition[key] = pos
	// Use plan id as index
	s.indexID[p.PlanID] = pos

	return s
}

func (s *stripeStore) findByEdition(key string) (PlanConfig, error) {
	i, ok := s.indexEdition[key]
	if !ok {
		return PlanConfig{}, fmt.Errorf("stripe plan for %s is not found", key)
	}

	return s.plans[i], nil
}

func (s *stripeStore) findByID(planID string) (PlanConfig, error) {
	i, ok := s.indexID[planID]
	if !ok {
		return PlanConfig{}, fmt.Errorf("stripe plan with id %s is not found", planID)
	}

	return s.plans[i], nil
}

var stripePlans = newStripeStoreSchema()

func GetPlanByEdition(key string, live bool) (PlanConfig, error) {
	return stripePlans.findByEdition(key + "_" + stripeKeySuffix[live])
}

func GetPlanByID(id string) (PlanConfig, error) {
	return stripePlans.findByID(id)
}
