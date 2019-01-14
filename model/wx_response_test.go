package model

import (
	"testing"

	"gitlab.com/ftchinese/subscription-api/wechat"
)

func TestSavePrepayResp(t *testing.T) {
	m := newMocker()
	subs := m.wxpaySubs()

	p, err := wxParsedPrepay()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Parsed response: %+v\n", p)

	uor := wechat.NewUnifiedOrderResp(p)
	t.Logf("UnifiedOrderResp: %+v\n", uor)

	err = devEnv.SavePrepayResp(subs.OrderID, uor)
	if err != nil {
		t.Error(err)
	}
}

func TestSaveWxNoti(t *testing.T) {
	m := newMocker()
	subs := m.wxpaySubs()

	p, err := wxParsedNoti(subs.OrderID)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Parsed notification: %+v\n", p)

	noti := wechat.NewNotification(p)
	t.Logf("Notification: %+v\n", noti)

	err = devEnv.SaveWxNotification(noti)
	if err != nil {
		t.Error(err)
	}
}

// Generate data to be used by Postman.
func TestGenerateNoti(t *testing.T) {
	// User must exist in database and the email must be real; otherwise email cannot be sent.
	m := newMocker().withEmail("neefrankie@163.com")

	user, err := m.createUser()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Created user: %+v\n", user)

	subs, err := m.createWxpaySubs()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Generated an order: %+v\n", subs)

	resp := wxNotiResp(subs.OrderID)
	t.Logf("Mock response: %s\n", resp)
}
