package controller

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/wepay"
)

var appID = os.Getenv("WXPAY_APPID")
var mchID = os.Getenv("WXPAY_MCHID")
var apiKey = os.Getenv("WXPAY_API_KEY")

var mockApp = wepay.PayApp{
	AppID:  appID,
	MchID:  mchID,
	APIKey: apiKey,
	IsProd: false,
}

var mockClient = wxpay.NewClient(wxpay.NewAccount(appID, mchID, apiKey, false))

var mockPlan = model.DefaultPlans["standard_year"]

func generateWxID() string {
	id, _ := util.RandomBase64(21)
	return id
}

func createNotification() string {
	p := make(wxpay.Params)

	nonce, _ := util.RandomHex(16)

	p.SetString("return_code", "SUCCESS").
		SetString("return_msg", "OK").
		SetString("appid", appID).
		SetString("mch_id", mchID).
		SetString("nonce_str", nonce).
		SetString("result_code", "SUCCESS").
		SetString("openid", generateWxID()).
		SetString("is_subscribe", "N").
		SetString("trade_type", "APP").
		SetString("bank_type", "CMC").
		SetString("total_fee", "25800").
		SetString("cash_fee", "25800").
		SetString("transaction_id", "1217752501201407033233368018").
		SetString("out_trade_no", mockPlan.OrderID()).
		SetString("time_end", time.Now().Format("20060102150405"))

	s := mockClient.Sign(p)

	signedParams := p.SetString("sign", s)

	return wxpay.MapToXml(signedParams)
}

func TestCreateNotification(t *testing.T) {
	n := createNotification()

	t.Logf("A mock notification: %s\n", n)
}

func TestProcessResponse(t *testing.T) {
	n := createNotification()
	resp := strings.NewReader(n)
	params, err := wepay.ParseResponse(mockClient, resp)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Parsed response: %+v\n", params)
}
