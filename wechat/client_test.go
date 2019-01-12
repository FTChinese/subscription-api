package wechat

import (
	"os"
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

func (m mocker) prepayResp() string {
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

func TestIsValidSign(t *testing.T) {
	order := GenerateUnifiedOrder(mockPlan, fake.IPv4(), mockPlan.OrderID())

	h := mockClient.Sign(order)

	order.SetString(wxpay.Sign, h)

	t.Logf("Unified order: %+v\n", order)

	ok := mockClient.ValidSign(order)

	t.Logf("Is signature valid: %t\n", ok)
}
