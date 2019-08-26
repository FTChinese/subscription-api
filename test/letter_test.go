package test

import (
	"gitlab.com/ftchinese/subscription-api/models/letter"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"testing"
)

func TestNewSubParcel(t *testing.T) {

	profile := NewProfile()

	account := profile.Account(reader.AccountKindFtc)

	store := NewSubStore(profile, reader.AccountKindFtc)

	order := store.MustCreate(YearlyStandardLive)
	order = store.MustConfirm(order.ID)

	parcel, err := letter.NewSubParcel(account, order)

	if err != nil {
		t.Error(err)
	}

	t.Logf("New Subscription parcel: %+v", parcel)
}

func TestRenewalParcel(t *testing.T) {
	profile := NewProfile()

	account := profile.Account(reader.AccountKindFtc)

	store := NewSubStore(profile, reader.AccountKindFtc)

	orders := store.MustRenewN(YearlyStandardLive, 2)

	parcel, err := letter.NewRenewalParcel(account, orders[1])

	if err != nil {
		t.Error(err)
	}

	t.Logf("New Subscription parcel: %+v", parcel)
}

func TestUpgradingParcel(t *testing.T) {
	profile := NewProfile()

	account := profile.Account(reader.AccountKindFtc)

	store := NewSubStore(profile, reader.AccountKindFtc)

	store.MustRenewN(YearlyStandardLive, 2)

	order := store.MustConfirm(store.MustCreate(YearlyPremiumLive).ID)

	parcel, err := letter.NewUpgradeParcel(
		account,
		order,
		store.UpgradePlan)

	if err != nil {
		t.Error(err)
	}

	t.Logf("New Subscription parcel: %+v", parcel)
}
