package test

import (
	"github.com/google/uuid"
	"testing"
)

func TestUUID(t *testing.T) {
	t.Logf("FTC ID: %s", uuid.New().String())
}

func TestGenerateNewUser(t *testing.T) {
	t.Logf("%+v", NewProfile())
}

func TestModel_CreateNewMember(t *testing.T) {
	model := NewModel()
	store := NewSubStore(NewProfile().AccountID(AccountKindFtc))

	model.CreateNewMember(store)

	t.Logf("Created a new order %s", store.GetLastOrder().ID)
	t.Logf("Created a new member %+v", store.Member)
}
