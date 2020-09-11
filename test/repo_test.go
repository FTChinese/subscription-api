package test

import (
	"testing"
)

func TestRepo_SaveAccount(t *testing.T) {
	p := NewPersona()

	err := NewRepo().SaveAccount(p.FtcAccount())

	if err != nil {
		t.Error(err)
	}
}

func TestRepo_SaveMembership(t *testing.T) {
	p := NewPersona()

	err := NewRepo().SaveMembership(p.Membership())

	if err != nil {
		t.Error(err)
	}
}

func TestRepo_SaveOrder(t *testing.T) {
	p := NewPersona()

	err := NewRepo().SaveOrder(p.CreateOrder())

	if err != nil {
		t.Error(err)
	}
}
