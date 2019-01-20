package model

import (
	"testing"
)

func TestConfirmationParcel(t *testing.T) {
	m := newMocker()

	user := m.user()
	subs := m.confirmedSubs()

	p, err := user.ComfirmationParcel(subs)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Parcel: %+v\n", p)
}

func TestSendEmail(t *testing.T) {
	m := newMocker().withEmail("neefrankie@163.com")

	user, err := m.createUser()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Created or updated a mock user: %+v\n", user)

	subs := m.confirmedSubs()
	t.Logf("A confimed subscription: %+v\n", subs)

	parcel, err := user.ComfirmationParcel(subs)

	if err != nil {
		t.Error(err)
		return
	}

	err = postman.Deliver(parcel)
	if err != nil {
		t.Error(err)
	}
}
