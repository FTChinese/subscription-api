package wepay

import (
	"database/sql"
	"os"
	"time"

	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
)

func newDevEnv() Env {
	db, err := sql.Open("mysql", "sampadm:secret@unix(/tmp/mysql.sock)/")

	if err != nil {
		panic(err)
	}

	return Env{DB: db}
}

var devEnv = newDevEnv()
var paywallEnv = paywall.Env{}

var appID = os.Getenv("WXPAY_APPID")
var mchID = os.Getenv("WXPAY_MCHID")
var apiKey = os.Getenv("WXPAY_API_KEY")

var mockApp = PayApp{
	AppID:  appID,
	MchID:  mchID,
	APIKey: apiKey,
	IsProd: false,
}

var mockClient = wxpay.NewClient(wxpay.NewAccount(appID, mchID, apiKey, false))

var mockPlan = paywallEnv.GetCurrentPricing()["standard_year"]

func generateWxID() string {
	id, _ := util.RandomBase64(21)
	return id
}

func createNotificationResponse() string {
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
