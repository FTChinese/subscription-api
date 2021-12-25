package stripe

import (
	"testing"
)

func Test_newPlanStore(t *testing.T) {
	store := newPriceEditionStore()

	if len(store.editions) != 6 {
		t.Error("not 6")
	}

	if len(store.indexEdition) != 6 {
		t.Error("not 6")
	}

	if len(store.indexID) != 6 {
		t.Error("not 6")
	}
}
