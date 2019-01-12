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

func generateWxID() string {
	id, _ := util.RandomBase64(21)
	return id
}

func (c Client) generatNoti() string {
	p := make(wxpay.Params)

	nonce, _ := util.RandomHex(16)

	p.SetString("return_code", "SUCCESS")
	p.SetString("return_msg", "OK")
	p.SetString("appid", c.appID)
	p.SetString("mch_id", c.mchID)
	p.SetString("nonce_str", nonce)
	p.SetString("result_code", "SUCCESS")
	p.SetString("openid", generateWxID())
	p.SetString("is_subscribe", "N")
	p.SetString("trade_type", "APP")
	p.SetString("bank_type", "CMC")
	p.SetString("total_fee", "25800")
	p.SetString("cash_fee", "25800")
	p.SetString("transaction_id", "1217752501201407033233368018")
	p.SetString("out_trade_no", mockPlan.OrderID())
	p.SetString("time_end", time.Now().Format("20060102150405"))

	s := c.Client.Sign(p)

	p.SetString("sign", s)

	return wxpay.MapToXml(p)
}

func TestSignature(t *testing.T) {
	order := GenerateUnifiedOrder(mockPlan, fake.IPv4(), mockPlan.OrderID())

	h := mockClient.Sign(order)

	t.Logf("Hash: %s\n", h)
}

func TestIsValidSign(t *testing.T) {
	order := GenerateUnifiedOrder(mockPlan, fake.IPv4(), mockPlan.OrderID())

	h := mockClient.Sign(order)

	order.SetString(wxpay.Sign, h)

	ok := mockClient.ValidSign(order)

	t.Logf("Is signature valid: %t\n", ok)
}
