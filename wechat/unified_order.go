package wechat

import (
	"fmt"
	"github.com/FTChinese/go-rest"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"time"
)

// UnifiedOrder contains the data sent to wechat to obtain
// a prepay data.
type UnifiedOrder struct {
	Body        string
	OrderID     string
	Price       int64
	IP          string
	CallbackURL string
	TradeType   TradeType
	ProductID   string
	OpenID      string
}

// ToParam turns UnifiedOrder data structure to a map
// which can be used by `wxpay` package.
func (o UnifiedOrder) ToParam() wxpay.Params {
	p := make(wxpay.Params)
	p.SetString("body", o.Body)
	p.SetString("out_trade_no", o.OrderID)
	p.SetInt64("total_fee", o.Price)
	p.SetString("spbill_create_ip", o.IP)
	p.SetString("notify_url", o.CallbackURL)
	// APP for native app
	// NATIVE for web site
	// JSAPI for web page opend inside wechat browser
	p.SetString("trade_type", o.TradeType.String())

	switch o.TradeType {
	case TradeTypeWeb:
		p.SetString("product_id", o.ProductID)

	case TradeTypeJSAPI:
		p.SetString("opendid", o.OpenID)
	}

	return p
}

// UnifiedOrderResp contains the response data from Wechat unified order.
type UnifiedOrderResp struct {
	WxResp
	TradeType null.String
	PrepayID  null.String
	CodeURL   null.String
}

// NewUnifiedOrderResp creates converts PrePay from a wxpay.Params type.
// Example response from Wechat:
// map[
// result_code:SUCCESS
// trade_type:APP
// sign:C7493936018971251931EADC03FE0B46
// prepay_id:wx131027225284604cf9f311763035575963
// return_code:SUCCESS
// return_msg:OK
// appid:wxacddf1c20516eb69
// mch_id:1504993271
// nonce_str:aOyCOfOvWZQZkRwp
// ]
func NewUnifiedOrderResp(p wxpay.Params) UnifiedOrderResp {
	r := UnifiedOrderResp{}

	r.Populate(p)

	if v, ok := p["trade_type"]; ok {
		r.TradeType = null.StringFrom(v)
	}

	if v, ok := p["prepay_id"]; ok {
		r.PrepayID = null.StringFrom(v)
	}

	// For native pay.
	if v, ok := p["code_url"]; ok {
		r.CodeURL = null.StringFrom(v)
	}

	return r
}

// ToPrepay creates a Preapay from unified order response
// and subscription order.
func (o UnifiedOrderResp) ToPrepay(subs paywall.Subscription) Prepay {
	nonce, _ := gorest.RandomHex(10)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	return Prepay{
		FtcOrderID: subs.OrderID,
		Price:      subs.ListPrice,
		ListPrice:  subs.ListPrice,
		NetPrice:   subs.NetPrice,
		AppID:      o.AppID.String,
		PartnerID:  o.MID.String,
		Package:    "Sign=WXPay",
		Nonce:      nonce,
		Timestamp:  timestamp,
	}
}
