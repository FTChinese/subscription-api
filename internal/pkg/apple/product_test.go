package apple

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"testing"
)

func Test_appleStore_findByEdition(t *testing.T) {
	store := newAppleStore()

	p, err := store.findByEdition(price.StdYearEdition)

	if err != nil {
		t.Error(err)
	}

	t.Logf("%v", p)
}

func Test_appleStore_findByID(t *testing.T) {
	store := newAppleStore()

	for _, p := range store.products {
		found, err := store.findByID(p.ID)
		if err != nil {
			t.Error(err)
		}

		t.Logf("Found price: %+v", found)
	}
}
