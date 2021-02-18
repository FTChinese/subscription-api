package apple

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_newAppleStore(t *testing.T) {
	store := newAppleStore()

	assert.Len(t, store.products, 3)

	assert.Len(t, store.indexEdition, 3)

	assert.Len(t, store.indexID, 3)
}

func Test_appleStore_findByEdition(t *testing.T) {
	store := newAppleStore()

	p, err := store.findByEdition(price.NewStdYearEdition())

	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, p.Tier, enum.TierStandard)
	assert.Equal(t, p.Cycle, enum.CycleYear)
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
