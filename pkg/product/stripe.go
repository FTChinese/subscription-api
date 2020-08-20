package product

import (
	"fmt"
)

type StripePlan struct {
	Edition
	PlanID string
	Live   bool
}

func stripeEditionKey(e Edition, live bool) string {
	var suffix string
	if live {
		suffix = "live"
	} else {
		suffix = "test"
	}

	return e.NamedKey() + "_" + suffix
}

type stripeStore struct {
	plans        []StripePlan
	indexEdition map[string]int
	indexID      map[string]int
}

func newStripeStoreSchema() *stripeStore {
	return &stripeStore{
		plans:        make([]StripePlan, 0),
		indexEdition: make(map[string]int),
		indexID:      make(map[string]int),
	}
}

func (s *stripeStore) Add(p StripePlan) *stripeStore {
	s.plans = append(s.plans, p)

	pos := len(s.plans) - 1
	s.indexEdition[stripeEditionKey(p.Edition, p.Live)] = pos
	s.indexID[p.PlanID] = pos

	return s
}

func (s *stripeStore) findByEdition(e Edition, live bool) (StripePlan, error) {
	key := stripeEditionKey(e, live)
	i, ok := s.indexEdition[key]
	if !ok {
		return StripePlan{}, fmt.Errorf("stripe plan for %s is not found", key)
	}

	return s.plans[i], nil
}

func (s *stripeStore) findByID(planID string) (StripePlan, error) {
	i, ok := s.indexID[planID]
	if !ok {
		return StripePlan{}, fmt.Errorf("stripe plan with id %s is not found", planID)
	}

	return s.plans[i], nil
}

var stripePlans = newStripeStoreSchema().
	Add(StripePlan{
		Edition: NewStdMonthEdition(),
		PlanID:  "plan_FXZYLOEbcvj5Tx",
		Live:    true,
	}).
	Add(StripePlan{
		Edition: NewStdMonthEdition(),
		PlanID:  "plan_FOdgPTznDwHU4i",
		Live:    false,
	}).
	Add(StripePlan{
		Edition: NewStdYearEdition(),
		PlanID:  "plan_FXZZUEDpToPlZK",
		Live:    true,
	}).
	Add(StripePlan{
		Edition: NewStdYearEdition(),
		PlanID:  "plan_FOdfeaqzczp6Ag",
		Live:    false,
	}).
	Add(StripePlan{
		Edition: NewPremiumEdition(),
		PlanID:  "plan_FXZbv1cDTsUKOg",
		Live:    true,
	}).
	Add(StripePlan{
		Edition: NewPremiumEdition(),
		PlanID:  "plan_FOde0uAr0V4WmT",
		Live:    false,
	})

func GetStripePlanByEdition(edition Edition, live bool) (StripePlan, error) {
	return stripePlans.findByEdition(edition, live)
}

func GetStripePlanByID(id string) (StripePlan, error) {
	return stripePlans.findByID(id)
}
