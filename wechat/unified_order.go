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
	case TradeTypeDesktop:
		p.SetString("product_id", o.ProductID)

	case TradeTypeJSAPI:
		p.SetString("openid", o.OpenID)
	}

	return p
}

// UnifiedOrderResp contains the response data from Wechat unified order.
type UnifiedOrderResp struct {
	Resp
	TradeType null.String
	PrepayID  null.String
	CodeURL   null.String
	MWebURL   null.String
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

	if v, ok := p["mweb_url"]; ok {
		r.MWebURL = null.StringFrom(v)
	}

	return r
}

// ToPrepay creates a Preapay from unified order response
// and subscription order.
func (o UnifiedOrderResp) ToLegacyAppPay(subs paywall.Subscription) LegacyAppPay {
	nonce, _ := gorest.RandomHex(10)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	return LegacyAppPay{
		FtcOrderID: subs.OrderID,
		Price:      subs.ListPrice,
		ListPrice:  subs.ListPrice,
		NetPrice:   subs.NetPrice,
		AppID:      o.AppID.String,
		PartnerID:  o.MID.String,
		PrepayID:   o.PrepayID.String,
		Package:    "Sign=WXPay",
		Nonce:      nonce,
		Timestamp:  timestamp,
	}
}

func (o UnifiedOrderResp) ToAppPay(subs paywall.Subscription) AppPay {
	p := AppPay{
		PartnerID: o.MID.String,
		PrepayID:  o.PrepayID.String,
		Package:   "Sign=WXPay",
		Nonce:     GenerateNonce(),
		Timestamp: GenerateTimestamp(),
	}

	p.FtcOrderID = subs.OrderID
	p.ListPrice = subs.ListPrice
	p.NetPrice = subs.NetPrice
	p.AppID = o.AppID.String

	return p
}

// ToWxBrowserPay turns unified order response to data
// required by JSAPI.
func (o UnifiedOrderResp) ToWxBrowserPay(subs paywall.Subscription) WxBrowserPay {
	nonce, _ := gorest.RandomHex(10)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	p := WxBrowserPay{

		Timestamp: timestamp,
		Nonce:     nonce,
		Package:   "prepay_id" + o.PrepayID.String,
		SignType:  "MD5",
	}

	p.FtcOrderID = subs.OrderID
	p.ListPrice = subs.ListPrice
	p.NetPrice = subs.NetPrice
	p.AppID = o.AppID.String

	return p
}

func (o UnifiedOrderResp) ToDesktopPay(subs paywall.Subscription) DesktopPay {
	p := DesktopPay{
		CodeURL: o.CodeURL.String,
	}

	p.FtcOrderID = subs.OrderID
	p.ListPrice = subs.ListPrice
	p.NetPrice = subs.NetPrice
	p.AppID = o.AppID.String

	return p
}

func (o UnifiedOrderResp) ToMobilePay(subs paywall.Subscription) MobilePay {
	p := MobilePay{
		MWebURL: o.MWebURL.String,
	}

	p.FtcOrderID = subs.OrderID
	p.ListPrice = subs.ListPrice
	p.NetPrice = subs.NetPrice
	p.AppID = o.AppID.String

	return p
}
