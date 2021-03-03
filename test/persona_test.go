package test

import (
	"github.com/FTChinese/subscription-api/pkg/apple"
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

	m := apple.NewMembership(p.AccountID(), sub)

	m = m.Sync()

	assert.NotZero(t, m.LegacyExpire)
	assert.NotZero(t, m.LegacyTier)

	t.Log(m.LegacyTier)
	t.Log(m.LegacyExpire)
}
