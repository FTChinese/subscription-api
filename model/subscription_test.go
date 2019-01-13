package model

import (
	"testing"
	"time"

	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

func TestNewSubs(t *testing.T) {
	m := NewMocker()

	subs := m.WxpaySubs()

	t.Logf("An order: %+v\n", subs)

	subs.GenerateOrderID()

	t.Logf("After called GenerateOrderID again: %+v\n", subs)
}

func TestIsSubsAllowed(t *testing.T) {
	m := NewMocker()

	subs := paywall.NewWxpaySubs(m.UserID, mockPlan, enum.EmailLogin)

	ok, err := devEnv.IsSubsAllowed(subs)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Is subscription allowed: %t\n", ok)
}
func TestSaveSubs(t *testing.T) {
	m := NewMocker()

	wxSubs, err := m.CreateWxpaySubs()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Created a Wechat subscription: %+v\n", wxSubs)

	aliSubs, err := m.CreateAlipaySubs()
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Created a Ali subscription: %+v\n", aliSubs)
}

func TestVerifyWxNoti(t *testing.T) {
	m := NewMocker()
	subs, err := m.CreateWxpaySubs()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Created a subscription order: %+v", subs)

	p, err := WxParsedNoti(subs.OrderID)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Notification: %+v", p)

	err = devEnv.VerifyWxNotification(p)
	if err != nil {
		t.Error(err)
	}
}

func TestFindSubs(t *testing.T) {
	m := NewMocker()

	subs, err := m.CreateWxpaySubs()

	found, err := devEnv.FindSubscription(subs.OrderID)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Found subscription: %+v\n", found)
}

func TestRenewMember(t *testing.T) {
	m := NewMocker()

	// First iteration creates a new subscription,
	// second iteration renew the membership.
	for i := 0; i < 2; i++ {
		subs, err := m.CreateMember()
		if err != nil {
			t.Error(err)
			break
		}

		t.Logf("Subscripiton for a membership: %+v\n", subs)
	}
}

func TestCreatemember_wxLogin(t *testing.T) {
	m := NewMocker()

	for i := 0; i < 2; i++ {
		subs, err := m.CreateMember()
		if err != nil {
			t.Error(err)
			break
		}

		t.Logf("Subscription for a membership: %+v\n", subs)
	}
}

func TestConfirmPayment(t *testing.T) {
	m := NewMocker()

	subs, err := m.CreateWxpaySubs()
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
