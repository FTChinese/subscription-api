package price

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_newPlanStore(t *testing.T) {
	store := newStripeEditions()

	assert.Len(t, store.editions, 6)
	assert.Len(t, store.indexEdition, 6)
	assert.Len(t, store.indexID, 6)
}
