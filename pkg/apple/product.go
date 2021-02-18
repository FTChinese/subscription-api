package apple

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/price"
)

type Product struct {
	price.Edition
	ID string
}

type appleStore struct {
	products     []Product
	indexEdition map[string]int
	indexID      map[string]int
}

func newAppleStore() appleStore {
	s := appleStore{
		products: []Product{
			{
				Edition: price.NewStdMonthEdition(),
				ID:      "com.ft.ftchinese.mobile.subscription.member.monthly",
			},
			{
				Edition: price.NewStdYearEdition(),
				ID:      "com.ft.ftchinese.mobile.subscription.member",
			},
			{
				Edition: price.NewPremiumEdition(),
				ID:      "com.ft.ftchinese.mobile.subscription.vip",
			},
		},
		indexEdition: make(map[string]int),
		indexID:      make(map[string]int),
	}

	for i, v := range s.products {
		s.indexEdition[v.NamedKey()] = i
		s.indexID[v.ID] = i
	}

	return s
}

func (s appleStore) findByEdition(e price.Edition) (Product, error) {
	i, ok := s.indexEdition[e.NamedKey()]
	if !ok {
		return Product{}, fmt.Errorf("apple price for %s is not found", e.NamedKey())
	}

	return s.products[i], nil
}

func (s appleStore) findByID(id string) (Product, error) {
	i, ok := s.indexID[id]
	if !ok {
		return Product{}, fmt.Errorf("apple prodct with id %s not found", id)
	}

	return s.products[i], nil
}

func (s appleStore) exists(id string) bool {
	_, ok := s.indexID[id]

	return ok
}

var appleProducts = newAppleStore()
