package apple

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/product"
)

type Product struct {
	product.Edition
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
				Edition: product.NewStdMonthEdition(),
				ID:      "com.ft.ftchinese.mobile.subscription.member.monthly",
			},
			{
				Edition: product.NewStdYearEdition(),
				ID:      "com.ft.ftchinese.mobile.subscription.member",
			},
			{
				Edition: product.NewPremiumEdition(),
				ID:      "com.ft.ftchinese.mobile.subscription.vip",
			},
		},
		indexEdition: make(map[string]int),
		indexID:      make(map[string]int),
	}

	for i, v := range s.products {
		s.indexEdition[v.NamedKey()] = i
		s.indexEdition[v.ID] = i
	}

	return s
}

func (s appleStore) findByEdition(e product.Edition) (Product, error) {
	i, ok := s.indexEdition[e.NamedKey()]
	if !ok {
		return Product{}, fmt.Errorf("apple product for %s is not found", e.NamedKey())
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

var appleProducts = newAppleStore()

func GetProductByEdition(e product.Edition) (Product, error) {
	return appleProducts.findByEdition(e)
}

func GetProductByID(id string) (Product, error) {
	return appleProducts.findByID(id)
}
