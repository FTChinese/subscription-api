package stripe

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_newPlanStore(t *testing.T) {
	store := newPlanStore()

	assert.Len(t, store.plans, 6)
	assert.Len(t, store.indexEdition, 6)
	assert.Len(t, store.indexID, 6)
}
