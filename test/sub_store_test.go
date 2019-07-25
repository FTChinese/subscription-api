package test

import (
	"gitlab.com/ftchinese/subscription-api/paywall"
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

	t.Logf("%+v", store)
}

func TestSubStore_GetLastOrder(t *testing.T) {
	store := NewSubStore(NewProfile().AccountID(AccountKindFtc))

	store.AddOrder(paywall.SubsKindCreate)
	store.AddOrder(paywall.SubsKindRenew)

	t.Logf("Store: %+v", store)
	t.Logf("Last order: %+v", store.GetLastOrder())
}
