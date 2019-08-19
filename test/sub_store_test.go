package test

import (
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"testing"
)

func TestNewSubStore(t *testing.T) {
	id := NewProfile().AccountID(AccountKindFtc)

	store := NewSubStore(id)

	t.Logf("%+v", store)
}

func TestSubStore_AddOrder(t *testing.T) {
	store := NewSubStore(NewProfile().AccountID(AccountKindFtc))

	store.AddOrder(paywall.SubsKindCreate)

	t.Logf("Current orders %+v", store.Orders)
}

func TestSubStore_GetOrder(t *testing.T) {
	store := NewSubStore(NewProfile().AccountID(AccountKindFtc))

	o := store.AddOrder(paywall.SubsKindCreate)

	got, err := store.GetOrder(o.ID)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Get order %+v", got)
}

func TestSubStore_ConfirmOrder(t *testing.T) {
	store := NewSubStore(NewProfile().FtcAccountID())

	o := store.AddOrder(paywall.SubsKindCreate)

	o, err := store.ConfirmOrder(o.ID)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Confirmed order: %+v", o)
}

func TestSubStore_UpgradeOrder(t *testing.T) {
	store := NewSubStore(NewProfile().FtcAccountID())

	upgradeOrder, err := store.UpgradeOrder(5)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Upgrade order: %+v", upgradeOrder)
	t.Logf("Upgrade v1: %+v", store.UpgradeV1)
	t.Logf("Upgrade v2: %+v", store.UpgradeV2)
}
