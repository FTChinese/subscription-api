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

// AppOrder creates an order used when your called wechat pay
// inside your own app.
type AppOrder struct {
	paywall.Subscription
	FtcOrderID string `json:"ftcOrderId"` // Deprecate
	AppID      string `json:"appId"`
	PartnerID  string `json:"partnerId"`
	PrepayID   string `json:"prepayId"`
	Timestamp  string `json:"timestamp"`
	Nonce      string `json:"nonce"`
	Package    string `json:"pkg"`
	Signature  string `json:"signature"`
}

func (a *AppOrder) Params() wxpay.Params {
	param := make(wxpay.Params)
	param["appid"] = a.WxAppID.String
	param["partnerid"] = a.PartnerID
	param["prepayid"] = a.PrepayID
	param["package"] = a.Package
	param["noncestr"] = a.Nonce
	param["timestamp"] = a.Timestamp

	return param
}

// BuildAppOrder creates an order that can be used inside
// a native app.
func (c Client) BuildAppOrder(u UnifiedOrderResp, subs paywall.Subscription) AppOrder {
	o := AppOrder{
		Subscription: subs,
		FtcOrderID:   subs.OrderID,
		AppID:        subs.WxAppID.String,
		PartnerID:    u.MID.String,
		PrepayID:     u.PrepayID.String,
		Timestamp:    GenerateTimestamp(),
		Nonce:        GenerateNonce(),
		Package:      "Sign=WXPay",
	}

	o.Signature = c.Sign(o.Params())

	return o
}

// InAppBrowserOrder creates an order when user is trying to pay
// inside wechat's embedded browser.
// This is actually similar to AppOrder since they are all
// perform actions inside wechat app.
// It's a shame wechat cannot even use the same data structure
// for such insignificant differences.
type InAppBrowserOrder struct {
	paywall.Subscription
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Package   string `json:"pkg"`
	Signature string `json:"signature"`
	SignType  string `json:"signType"`
}

func (w *InAppBrowserOrder) Params() wxpay.Params {
	param := make(wxpay.Params)
	param["appId"] = w.WxAppID.String
	param["timeStamp"] = w.Timestamp
	param["nonceStr"] = w.Nonce
	param["package"] = w.Package
	param["signType"] = w.SignType

	return param
}

// BuildInAppBrowserOrder creates an order for payment inside
// wechat embedded browser.
func (c Client) BuildInAppBrowserOrder(u UnifiedOrderResp, subs paywall.Subscription) InAppBrowserOrder {
	o := InAppBrowserOrder{
		Subscription: subs,
		Timestamp:    GenerateTimestamp(),
		Nonce:        GenerateNonce(),
		Package:      "prepay_id=" + u.PrepayID.String,
		SignType:     "MD5",
	}

	o.Signature = c.Sign(o.Params())

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
