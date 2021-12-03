package stripe

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
)

// PriceEdition maps price.Edition to stripe price id.
type PriceEdition struct {
	price.Edition
	PriceID string
	Live    bool
}

type priceEditionStore struct {
	editions     []PriceEdition
	indexEdition map[string]int // Index the editions array by `<tier>_<cycle>_<live | test>`
	indexID      map[string]int // Index the editions array by stripe plan/price id.
}

func newPriceEditionStore() *priceEditionStore {
	s := &priceEditionStore{
		editions: []PriceEdition{
			{
				Edition: price.StdMonthEdition,
				PriceID: "price_1IM2Z4BzTK0hABgJ9Sh0u35h",
				Live:    true,
			},
			{
				Edition: price.StdMonthEdition,
				PriceID: "price_1IM2mgBzTK0hABgJVH8o9Sjm",
				Live:    false,
			},
			{
				Edition: price.StdYearEdition,
				PriceID: "price_1IM2aNBzTK0hABgJeJVIx3kL",
				Live:    true,
			},
			{
				Edition: price.StdYearEdition,
				PriceID: "price_1IM2nFBzTK0hABgJiIDeDIox",
				Live:    false,
			},
			{
				Edition: price.PremiumEdition,
				PriceID: "plan_FXZbv1cDTsUKOg",
				Live:    true,
			},
			{
				Edition: price.PremiumEdition,
				PriceID: "plan_FOde0uAr0V4WmT",
				Live:    false,
			},
		},
		indexEdition: make(map[string]int),
		indexID:      make(map[string]int),
	}

	for i, v := range s.editions {
		key := v.NamedKey() + "_" + ids.GetBoolKey(v.Live)
		s.indexEdition[key] = i
		s.indexID[v.PriceID] = i
	}

	return s
}

func (s priceEditionStore) FindByEdition(e price.Edition, live bool) (PriceEdition, error) {
	i, ok := s.indexEdition[e.NamedKey()+"_"+ids.GetBoolKey(live)]
	if !ok {
		return PriceEdition{}, fmt.Errorf("stripe price for %s is not found", e)
	}

	return s.editions[i], nil
}

func (s priceEditionStore) MustFindByEdition(e price.Edition, live bool) PriceEdition {
	se, err := s.FindByEdition(e, live)
	if err != nil {
		panic(err)
	}

	return se
}

// FindByID gets PriceEdition by stripe price id.
func (s priceEditionStore) FindByID(priceID string) (PriceEdition, error) {
	i, ok := s.indexID[priceID]
	if !ok {
		return PriceEdition{}, fmt.Errorf("stripe plan with id %s is not found", priceID)
	}

	return s.editions[i], nil
}

// PriceEditionStore keeps a copy of stripe price ids.
// Deprecated.
var PriceEditionStore = newPriceEditionStore()
