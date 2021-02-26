package price

import (
	"fmt"
)

var editionKeySuffix = map[bool]string{
	true:  "live",
	false: "test",
}

// StripeEdition contains ftc associates ftc price edition with Stripe
// plan/price id.
type StripeEdition struct {
	Edition
	PriceID string
	Live    bool
}

type stripeEditions struct {
	editions     []StripeEdition
	indexEdition map[string]int // Index the editions array by `<tier>_<cycle>_<live | test>`
	indexID      map[string]int // Index the editions array by stripe plan/price id.
}

func newStripeEditions() *stripeEditions {
	s := &stripeEditions{
		editions: []StripeEdition{
			{
				Edition: StdMonthEdition,
				PriceID: "price_1IM2Z4BzTK0hABgJ9Sh0u35h",
				Live:    true,
			},
			{
				Edition: StdMonthEdition,
				PriceID: "price_1IM2mgBzTK0hABgJVH8o9Sjm",
				Live:    false,
			},
			{
				Edition: StdYearEdition,
				PriceID: "price_1IM2aNBzTK0hABgJeJVIx3kL",
				Live:    true,
			},
			{
				Edition: StdYearEdition,
				PriceID: "price_1IM2nFBzTK0hABgJiIDeDIox",
				Live:    false,
			},
			{
				Edition: PremiumEdition,
				PriceID: "plan_FXZbv1cDTsUKOg",
				Live:    true,
			},
			{
				Edition: PremiumEdition,
				PriceID: "plan_FOde0uAr0V4WmT",
				Live:    false,
			},
		},
		indexEdition: make(map[string]int),
		indexID:      make(map[string]int),
	}

	for i, v := range s.editions {
		key := v.NamedKey() + "_" + editionKeySuffix[v.Live]
		s.indexEdition[key] = i
		s.indexID[v.PriceID] = i
	}

	return s
}

func (s stripeEditions) FindByEdition(e Edition, live bool) (StripeEdition, error) {
	i, ok := s.indexEdition[e.NamedKey()+"_"+editionKeySuffix[live]]
	if !ok {
		return StripeEdition{}, fmt.Errorf("stripe plan for %s is not found", e)
	}

	return s.editions[i], nil
}

func (s stripeEditions) MustFindByEdition(e Edition, live bool) StripeEdition {
	se, err := s.FindByEdition(e, live)
	if err != nil {
		panic(err)
	}

	return se
}

// FindByID gets StripeEdition by stripe price id.
func (s stripeEditions) FindByID(priceID string) (StripeEdition, error) {
	i, ok := s.indexID[priceID]
	if !ok {
		return StripeEdition{}, fmt.Errorf("stripe plan with id %s is not found", priceID)
	}

	return s.editions[i], nil
}

var StripeEditions = newStripeEditions()
