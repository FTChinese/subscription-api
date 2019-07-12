package test

import (
	"gitlab.com/ftchinese/subscription-api/paywall"
	"testing"
)

func TestMembership_FromAliOrWx(t *testing.T) {
	p := NewProfile()
	order := SubsConfirmed(p.RandomUserID())

	m, err := paywall.Membership{}.FromAliOrWx(order)
	if err != nil {
		t.Error(err)
	}

	t.Logf("New Membership: %+v", m)

	m, err = m.FromAliOrWx(order)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Renewed membership: %+v", m)
}
