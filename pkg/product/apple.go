package product

import "fmt"

type AppleProduct struct {
	Edition
	ID string
}

type appleStore struct {
	products     []AppleProduct
	indexEdition map[string]int
	indexID      map[string]int
}

func newAppleStore() appleStore {
	s := appleStore{
		products: []AppleProduct{
			{
				Edition: NewStdMonthEdition(),
				ID:      "com.ft.ftchinese.mobile.subscription.member.monthly",
			},
			{
				Edition: NewStdYearEdition(),
				ID:      "com.ft.ftchinese.mobile.subscription.member",
			},
			{
				Edition: NewPremiumEdition(),
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

func (s appleStore) findByEdition(e Edition) (AppleProduct, error) {
	i, ok := s.indexEdition[e.NamedKey()]
	if !ok {
		return AppleProduct{}, fmt.Errorf("apple product for %s is not found", e.NamedKey())
	}

	return s.products[i], nil
}

func (s appleStore) findByID(id string) (AppleProduct, error) {
	i, ok := s.indexID[id]
	if !ok {
		return AppleProduct{}, fmt.Errorf("apple prodct with id %s not found", id)
	}

	return s.products[i], nil
}

var appleProducts = newAppleStore()

func GetAppleProductByEdition(e Edition) (AppleProduct, error) {
	return appleProducts.findByEdition(e)
}

func GetAppleProductByID(id string) (AppleProduct, error) {
	return appleProducts.findByID(id)
}
