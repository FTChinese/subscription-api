package test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewProfile(t *testing.T) {
	t.Log(NewPersona().Email)
	t.Log(NewPersona().Email)
}

func TestPersona_IAPSubs(t *testing.T) {
	p := NewPersona()

	sub := p.IAPSubs()

	m := sub.NewMembership(p.AccountID())

	m = m.Sync()

	assert.NotZero(t, m.LegacyExpire)
	assert.NotZero(t, m.LegacyTier)

	t.Log(m.LegacyTier)
	t.Log(m.LegacyExpire)
}
