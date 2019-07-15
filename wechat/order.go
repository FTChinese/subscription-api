package wechat

import (
	"fmt"
	"github.com/FTChinese/go-rest"
	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"time"
)

func GenerateNonce() string {
	nonce, _ := gorest.RandomHex(10)

	return nonce
}

func GenerateTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

type AppOrderParams struct {
	AppID     string `json:"appId"`
	PartnerID string `json:"partnerId"`
	PrepayID  string `json:"prepayId"`
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Package   string `json:"pkg"`
	Signature string `json:"signature"`
}

func (a AppOrderParams) ToMap() wxpay.Params {
	param := make(wxpay.Params)
	param["appid"] = a.AppID
	param["partnerid"] = a.PartnerID
	param["prepayid"] = a.PrepayID
	param["package"] = a.Package
	param["noncestr"] = a.Nonce
	param["timestamp"] = a.Timestamp

	return param
}

// AppOrder creates an order used when your called wechat pay
// inside your own app.
type AppOrder struct {
	paywall.Subscription
	FtcOrderID     string         `json:"ftcOrderId"` // Deprecate
	AppOrderParams                // Deprecate
	Params         AppOrderParams `json:"params"`
}

// BuildAppOrder creates an order that can be used inside
// a native app.
func (c Client) BuildAppOrder(u UnifiedOrderResp, subs paywall.Subscription) AppOrder {
	p := AppOrderParams{
		AppID:     subs.WxAppID.String,
		PartnerID: u.MID.String,
		PrepayID:  u.PrepayID.String,
		Timestamp: GenerateTimestamp(),
		Nonce:     GenerateNonce(),
		Package:   "Sign=WXPay",
	}
	p.Signature = c.Sign(p.ToMap())

	o := AppOrder{
		Subscription:   subs,
		FtcOrderID:     subs.OrderID,
		AppOrderParams: p,
		Params:         p,
	}
	return o
}

type InWxBrowserParams struct {
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Package   string `json:"pkg"`
	Signature string `json:"signature"`
	SignType  string `json:"signType"`
}

func (w InWxBrowserParams) ToMap(appID string) wxpay.Params {
	param := make(wxpay.Params)
	param["appId"] = appID
	param["timeStamp"] = w.Timestamp
	param["nonceStr"] = w.Nonce
	param["package"] = w.Package
	param["signType"] = w.SignType

	return param
}

// InAppBrowserOrder creates an order when user is trying to pay
// inside wechat's embedded browser.
// This is actually similar to AppOrder since they are all
// perform actions inside wechat app.
// It's a shame wechat cannot even use the same data structure
// for such insignificant differences.
type InAppBrowserOrder struct {
	paywall.Subscription
	InWxBrowserParams
	Params InWxBrowserParams
}

// BuildInAppBrowserOrder creates an order for payment inside
// wechat embedded browser.
func (c Client) BuildInAppBrowserOrder(u UnifiedOrderResp, subs paywall.Subscription) InAppBrowserOrder {

	p := InWxBrowserParams{
		Timestamp: GenerateTimestamp(),
		Nonce:     GenerateNonce(),
		Package:   "prepay_id=" + u.PrepayID.String,
		SignType:  "MD5",
	}
	p.Signature = c.Sign(p.ToMap(subs.WxAppID.String))

	o := InAppBrowserOrder{
		Subscription:      subs,
		InWxBrowserParams: p,
		Params:            p,
	}

	//o.Signature = c.Sign(o.Params())

	return o
}

// BrowserOrder creates order for browsers.
// RedirectUrl is added later so that we could merge BrowserOrder
// and MobileOrder into a single data structure.
// CodeURL and MWebURL are kept for backward compatibility.
// Once the app ftacademy-node finished migration, they will be
// removed.
type BrowserOrder struct {
	paywall.Subscription
	CodeURL     string `json:"codeUrl,omitempty"` // Deprecate
	MWebURL     string `json:"mWebUrl,omitempty"` //Deprecate
	RedirectURL string `json:"redirectUrl"`
}

func BuildDesktopOrder(o UnifiedOrderResp, subs paywall.Subscription) BrowserOrder {
	return BrowserOrder{
		Subscription: subs,
		CodeURL:      o.CodeURL.String,
		RedirectURL:  o.CodeURL.String,
	}
}

func BuildMobileOrder(o UnifiedOrderResp, subs paywall.Subscription) BrowserOrder {
	return BrowserOrder{
		Subscription: subs,
		MWebURL:      o.MWebURL.String,
		RedirectURL:  o.MWebURL.String,
	}
}
