package wechat

import (
	"fmt"
	"github.com/FTChinese/go-rest"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"time"
)

func GenerateNonce() string {
	nonce, _ := gorest.RandomHex(10)

	return nonce
}

func GenerateTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

// OrderReq contains the data sent to wechat api to create an order.
type OrderReq struct {
	Body          string      `xml:"body"`
	SellerOrderID string      `xml:"out_trade_no"`
	TotalAmount   int64       `xml:"total_fee"`
	WebhookURL    string      `xml:"notify_url"`
	ProductID     string      `xml:"product_id"`
	TxKind        TradeType   `xml:"trade_type"`
	UserIP        string      `xml:"spbill_create_ip"`
	OpenID        null.String `xml:"open_id"` // Required for JSAPI
}

// ToMap turns OrderReq struct to the map required by sdk.
func (o OrderReq) ToMap() wxpay.Params {
	p := make(wxpay.Params)
	p.SetString("body", o.Body)
	p.SetString("out_trade_no", o.SellerOrderID)
	p.SetInt64("total_fee", o.TotalAmount)
	p.SetString("spbill_create_ip", o.UserIP)
	p.SetString("notify_url", o.WebhookURL)
	// APP for native app
	// NATIVE for web site
	// JSAPI for web page opend inside wechat browser
	p.SetString("trade_type", o.TxKind.String())
	p.SetString("product_id", o.ProductID)

	if o.TxKind == TradeTypeJSAPI {
		p.SetString("openid", o.OpenID.String)
	}

	return p
}

// OrderResp is an payment intent create at wechat.
type OrderResp struct {
	BaseResp
	TradeType  null.String `db:"trade_type"`
	PrepayID   null.String `db:"prepay_id"`
	QRCode     null.String `db:"qr_code"`
	MWebURL    null.String `db:"mobile_redirect_url"`
	FtcOrderID string      `db:"order_id"`
}

// NewOrderResp creates converts PrePay from a wxpay.Params type.
// Example response from Wechat:
// map[
// result_code:SUCCESS
// trade_type:APP
// sign:C7493936018971251931EADC03FE0B46
// prepay_id:wx131027225284604cf9f311763035575963
// return_code:SUCCESS
// return_msg:OK
// appid:***REMOVED***
// mch_id:1504993271
// nonce_str:aOyCOfOvWZQZkRwp
// ]
func NewOrderResp(orderID string, p wxpay.Params) OrderResp {
	r := OrderResp{
		BaseResp:   NewBaseResp(p),
		FtcOrderID: orderID,
	}

	v, ok := p["trade_type"]
	r.TradeType = null.NewString(v, ok)

	v, ok = p["prepay_id"]
	r.PrepayID = null.NewString(v, ok)

	// For native pay.
	v, ok = p["code_url"]
	r.QRCode = null.NewString(v, ok)

	v, ok = p["mweb_url"]
	r.MWebURL = null.NewString(v, ok)

	return r
}

type NativeAppParams struct {
	AppID     string `json:"appId"`
	PartnerID string `json:"partnerId"`
	PrepayID  string `json:"prepayId"`
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Package   string `json:"pkg"`
	Signature string `json:"signature"`
}

func NewNativeAppParams(or OrderResp) NativeAppParams {
	return NativeAppParams{
		AppID:     or.AppID.String,
		PartnerID: or.MID.String,
		PrepayID:  or.PrepayID.String,
		Timestamp: GenerateTimestamp(),
		Nonce:     GenerateNonce(),
		Package:   "Sign=WXPay",
	}
}

func (p NativeAppParams) ToMap() wxpay.Params {
	return wxpay.Params{
		"appid":     p.AppID,
		"partnerid": p.PartnerID,
		"prepayid":  p.PrepayID,
		"package":   p.Package,
		"noncestr":  p.Nonce,
		"timestamp": p.Timestamp,
	}
}

// InAppBrowserOrder creates an order when user is trying to pay
// inside wechat's embedded browser.
// This is actually similar to AppOrder since they are all
// perform actions inside wechat app.
// It's a shame wechat cannot even use the same data structure
// for such insignificant differences.
type JSApiParams struct {
	AppID     string `json:"appId"`
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Package   string `json:"pkg"`
	Signature string `json:"signature"`
	SignType  string `json:"signType"`
}

// NewJSApiParams creates a new JSApiParams to be signed.
func NewJSApiParams(or OrderResp) JSApiParams {
	return JSApiParams{
		AppID:     or.AppID.String,
		Timestamp: GenerateTimestamp(),
		Nonce:     GenerateNonce(),
		Package:   "prepay_id=" + or.PrepayID.String,
		SignType:  "MD5",
	}
}

// ToMap turns struct to a map so that we could generate signature from sdk.
func (p JSApiParams) ToMap() wxpay.Params {
	return wxpay.Params{
		"appId":     p.AppID,
		"timeStamp": p.Timestamp,
		"nonceStr":  p.Nonce,
		"package":   p.Package,
		"signType":  p.SignType,
	}
}
