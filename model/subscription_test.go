package model

import (
	"testing"
	"time"

	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

func TestCreateUser(t *testing.T) {
	m := newMocker()

	user, err := m.createUser()
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Created user: %+v\n", user)
}

func TestNewSubs(t *testing.T) {
	m := newMocker()
	wxSubs := m.wxpaySubs()
	t.Logf("Wxpay subscription with email login: %+v\n", wxSubs)

	aliSubs := m.alipaySubs()
	t.Logf("Alipay subscription with email login: %+v\n", aliSubs)

	m = m.withWxLogin()
	wxSubs = m.wxpaySubs()
	t.Logf("Wxpay subscription with wechat login: %+v\n", wxSubs)

	aliSubs = m.alipaySubs()
	t.Logf("Alipay subscription with wechat login: %+v\n", aliSubs)
}

func TestIsSubsAllowed(t *testing.T) {
	m := newMocker()

	subs := paywall.NewWxpaySubs(m.userID, mockPlan, enum.LoginMethodWx)

	ok, err := devEnv.IsSubsAllowed(subs)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Is subscription allowed: %t\n", ok)
}
func TestSaveSubs_emailLogin(t *testing.T) {
	m := newMocker()

	wxSubs, err := m.createWxpaySubs()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Created a Wechat subscription: %+v\n", wxSubs)

	aliSubs, err := m.createAlipaySubs()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Created a Ali subscription: %+v\n", aliSubs)
}

func TestSaveSubs_wxLogin(t *testing.T) {
	m := newMocker().withWxLogin()

	wxSubs, err := m.createWxpaySubs()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Created wxpay subscription with wechat login: %+v\n", wxSubs)

	aliSubs, err := m.createAlipaySubs()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Created alipay subscription with wechat login: %+v\n", aliSubs)
}

func TestFindSubs(t *testing.T) {
	m := newMocker()

	subs, err := m.createWxpaySubs()

	found, err := devEnv.FindSubscription(subs.OrderID)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Found subscription: %+v\n", found)
}

func TestConfirmPayment(t *testing.T) {
	m := newMocker()

	subs, err := m.createWxpaySubs()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Wxpay subscription with email login: %+v\n", subs)

	subs, err = devEnv.ConfirmPayment(subs.OrderID, time.Now())
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Confirmed subscription: %+v\n", subs)
}

func TestMultiConfirmPayment(t *testing.T) {
	m := newMocker()

	subs, err := m.createMember()

	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Confirmed subscription: %+v\n", subs)

	t.Log("Confirm again")

	subs, err = devEnv.ConfirmPayment(subs.OrderID, time.Now())
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Re-confirmed: %+v\n", subs)
}

func TestCreateMember_emailLogin(t *testing.T) {
	m := newMocker()

	user, err := m.createUser()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Created an FTC user: %+v\n", user)

	subs, err := m.createMember()
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Created a member with email login: %+v\n", subs)
}
func TestCreateMember_wxLogin(t *testing.T) {
	m := newMocker().withWxLogin()

	wxUser, err := m.createWxUser()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Created a wechat user: %+v\n", wxUser)

	subs, err := m.createMember()
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Created a member with wechat login: %+v\n", subs)
}

func TestFindMember(t *testing.T) {
	m := newMocker()

	subs, err := m.createMember()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Subscription: %+v\n", subs)

	member, err := devEnv.findMember(subs)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Found membership: %+v\n", member)
}

func TestFindDuration(t *testing.T) {
	m := newMocker()

	subs, err := m.createMember()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Subscription: %+v\n", subs)

	var dur paywall.Duration
	err = devEnv.db.QueryRow(
		devEnv.stmtSelectExpLock(subs.IsWxLogin()),
		subs.UserID,
	).Scan(
		&dur.Timestamp,
		&dur.ExpireDate,
	)

	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Found membership duration: %+v\n", dur)
}

func TestRenewMember(t *testing.T) {
	m := newMocker()

	for i := 0; i < 2; i++ {
		subs, err := m.createMember()
		if err != nil {
			t.Error(err)
			break
		}
		t.Logf("Subscritpion %d: %+v\n", i, subs)
	}
}
