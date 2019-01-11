package model

import (
	"testing"
	"time"

	"gitlab.com/ftchinese/subscription-api/enum"
)

func TestCreateSubs_emailLogin(t *testing.T) {
	user := NewUser()
	wxSubs, err := user.CreateWxpaySubs()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Created a Wechat subscription: %+v\n", wxSubs)

	aliSubs, err := user.CreateAlipaySubs()
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Created a Ali subscription: %+v\n", aliSubs)
}

func TestIsSubsAllowed(t *testing.T) {
	user := NewUser()

	subs, err := user.CreateWxpaySubs()

	subs = NewWxpaySubs(user.UserID, mockPlan, enum.EmailLogin)

	ok, err := devEnv.IsSubsAllowed(subs)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Is subscription allowed: %t\n", ok)
}

func TestFindSubs(t *testing.T) {
	user := NewUser()

	subs, err := user.CreateWxpaySubs()

	found, err := devEnv.FindSubscription(subs.OrderID)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Found subscription: %+v\n", found)
}

func TestRenewMember(t *testing.T) {
	user := NewUser()

	// First iteration creates a new subscription,
	// second iteration renew the membership.
	for i := 0; i < 2; i++ {
		subs, err := user.CreateMember()
		if err != nil {
			t.Error(err)
			break
		}

		t.Logf("Subscripiton for a membership: %+v\n", subs)
	}
}

func TestCreatemember_wxLogin(t *testing.T) {
	user := NewUser()

	for i := 0; i < 2; i++ {
		subs, err := user.CreateMember()
		if err != nil {
			t.Error(err)
			break
		}

		t.Logf("Subscription for a membership: %+v\n", subs)
	}
}

func TestConfirmPayment(t *testing.T) {
	user := NewUser()
	subs, err := user.CreateWxpaySubs()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("User placed a new order: %+v\n", subs)

	subs, err = devEnv.ConfirmPayment(subs.OrderID, time.Now())
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Confirmed subscription: %+v\n", subs)
}
