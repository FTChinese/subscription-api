package wechat

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"

	"github.com/icrowley/fake"
)

var appID = os.Getenv("WXPAY_APPID")
var mchID = os.Getenv("WXPAY_MCHID")
var apiKey = os.Getenv("WXPAY_API_KEY")

var mockClient = NewClient(appID, mchID, apiKey)

var mockPlan = paywall.GetDefaultPricing()["standard_year"]

// Mock a user for a specific subscription session.
type mocker struct {
	orderID string
	openID  string
	ip      string
}

func newMocker() mocker {
	id, _ := util.RandomBase64(21)

	return mocker{
		orderID: mockPlan.OrderID(),
		openID:  id,
		ip:      fake.IPv4(),
	}
}

func (m mocker) unifiedOrder() wxpay.Params {
	return GenerateUnifiedOrder(mockPlan, m.ip, m.orderID)
}

func fillResp(p wxpay.Params) wxpay.Params {
	nonce, _ := util.RandomHex(16)

	p.SetString("return_code", "SUCCESS")
	p.SetString("return_msg", "OK")
	p.SetString("appid", appID)
	p.SetString("mch_id", mchID)
	p.SetString("nonce_str", nonce)
	p.SetString("result_code", "SUCCESS")
	p.SetString("trade_type", "APP")

	return p
}

func (m mocker) uniOrderResp() string {
	p := make(wxpay.Params)

	p = fillResp(p)
	p.SetString("prepay_id", fake.CharactersN(26))

	s := mockClient.Sign(p)

	p.SetString("sign", s)

	return wxpay.MapToXml(p)
}

func (m mocker) notiResp() string {
	p := make(wxpay.Params)

	p = fillResp(p)
	p.SetString("openid", m.openID)
	p.SetString("is_subscribe", "N")
	p.SetString("bank_type", "CMC")
	p.SetString("total_fee", "25800")
	p.SetString("cash_fee", "25800")
	p.SetString("transaction_id", "1217752501201407033233368018")
	p.SetString("out_trade_no", mockPlan.OrderID())
	p.SetString("time_end", time.Now().Format("20060102150405"))

	s := mockClient.Sign(p)

	p.SetString("sign", s)

	return wxpay.MapToXml(p)
}

func TestUnifiedOrder(t *testing.T) {
	m := newMocker()

	uo := m.unifiedOrder()

	t.Logf("Created unified order: %+v\n", uo)
}
func TestSignature(t *testing.T) {
	m := newMocker()

	uo := m.unifiedOrder()

	t.Logf("Created unified order: %+v\n", uo)

	h := mockClient.Sign(uo)

	t.Logf("Hash: %s\n", h)
}

func TestSendUnifiedOrder(t *testing.T) {
	m := newMocker()
	uo := m.unifiedOrder()

	resp, err := mockClient.UnifiedOrder(uo)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Unified order response: %+v\n", resp)
}

func TestUnifiedOrderResp(t *testing.T) {
	m := newMocker()
	resp := m.uniOrderResp()

	t.Logf("Mock unified order response: %+v\n", resp)
}

func TestVerifyIdentity(t *testing.T) {
	m := newMocker()
	resp := m.uniOrderResp()

	p, err := mockClient.ParseResponse(strings.NewReader(resp))

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Response: %+v\n", p)

	r := mockClient.ValidateResponse(p)

	if r != nil {
		t.Error("Identity not verified.")
	}
}

func TestIsValidSign(t *testing.T) {
	m := newMocker()

	resp := m.uniOrderResp()

	p, err := mockClient.ParseResponse(strings.NewReader(resp))

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Response: %+v\n", p)

	ok := mockClient.ValidSign(p)

	if !ok {
		t.Error("Signature wrong")
		return
	}

	t.Logf("Is signature valid: %t\n", ok)
}

func TestNotiResp(t *testing.T) {
	m := newMocker()
	resp := m.notiResp()

	t.Logf("Mock notification: %+v\n", resp)

	p, err := mockClient.ParseResponse(strings.NewReader(resp))
	if err != nil {
		t.Error(err)
		return
	}

	noti := NewNotification(p)

	t.Logf("Notification: %+v\n", noti)
}
