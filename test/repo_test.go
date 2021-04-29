package test

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"testing"
)

func TestRepo_SaveMembership(t *testing.T) {
	p := NewPersona()

	err := NewRepo().SaveMembership(p.Membership())

	if err != nil {
		t.Error(err)
	}
}

func TestRepo_SaveOrder(t *testing.T) {
	p := NewPersona()

	err := NewRepo().SaveOrder(p.NewOrder(enum.OrderKindCreate))

	if err != nil {
		t.Error(err)
	}
}

func TestRepo_SaveIAPSubs(t *testing.T) {
	p := NewPersona()

	err := NewRepo().SaveIAPSubs(p.IAPSubs())
	if err != nil {
		t.Error(err)
	}
}

// Generate a mocking wx user
func TestRepo_SaveWxUser(t *testing.T) {
	unionID := faker.GenWxID()
	t.Logf("Wx user %s", unionID)

	schema := wxlogin.MockUserInfo(unionID)
	err := NewRepo().SaveWxUser(schema)
	if err != nil {
		t.Error(err)
	}
}
