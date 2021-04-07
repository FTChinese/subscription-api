package test

import (
	"github.com/FTChinese/go-rest/enum"
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
