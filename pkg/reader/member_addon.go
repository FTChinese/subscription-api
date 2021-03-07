package reader

import (
	"errors"
	"github.com/FTChinese/subscription-api/pkg/addon"
)

func (m Membership) WithAddOn(addOn addon.AddOn) Membership {
	m.AddOn = m.AddOn.Plus(addOn)
	return m
}

func (m Membership) CarriedOver() addon.AddOn {
	return addon.New(m.Tier, m.RemainingDays())
}

func (m Membership) HasAddOns() bool {
	return m.Standard > 0 || m.Premium > 0
}

func (m Membership) ShouldUseAddOn() error {
	if m.IsZero() {
		return errors.New("subscription backup days only applicable to an existing membership")
	}

	if !m.IsExpired() {
		return errors.New("backup days come into effect only after current subscription expired")
	}

	if !m.HasAddOns() {
		return errors.New("current membership does not have backup days")
	}

	return nil
}
