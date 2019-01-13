package model

import (
	"testing"

	"gitlab.com/ftchinese/subscription-api/wechat"
)

func TestSavePrepayResp(t *testing.T) {
	m := NewMocker()
	subs := m.WxpaySubs()

	p, err := MockParsedPrepay()
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
	m := NewMocker()
	subs := m.WxpaySubs()

	p, err := WxParsedNoti(subs.OrderID)
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

// Use the data the test Wechat response URL.
func TestGenerateNoti(t *testing.T) {
	m := NewMocker()

	subs, err := m.CreateWxpaySubs()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Generated an order: %+v\n", subs)

	resp := WxNotiResp(subs.OrderID)
	t.Logf("Mock response: %s\n", resp)
}
