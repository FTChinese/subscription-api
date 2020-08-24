package subs

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/wechat"
)

// WxpayNativeAppOrder creates an order used by native apps.
type WxpayNativeAppOrder struct {
	Order
	wechat.AppOrderParams
	Params wechat.AppOrderParams `json:"params"`
}

// WepayEmbedBrowserOrder responds to purchase made in wechat
// embedded browser.
// This is actually similar to AppOrder since they are all
// perform actions inside wechat app.
// It's a shame wechat cannot even use the same data structure
// for such insignificant differences.
type WxpayEmbedBrowserOrder struct {
	Order
	Params wechat.InWxBrowserParams `json:"params"`
}

// WxpayBrowserOrder creates order for payment via wechat
// made in browsers.
// For desktop browser, wechat send back a custom url
// for the client to generate a QR image;
// For mobile browser, wechat sends back a canonical url
// that can be redirected to.
// and MobileOrder into a single data structure.
type WxpayBrowserOrder struct {
	Order
	// TODO: rename json tag codeUrl to qrCode
	QRCode  string `json:"qrCodeUrl,omitempty"`         // Used by desktop browser. It is a custom url like wexin://wxpay/bizpayurl
	MWebURL string `json:"mobileRedirectUrl,omitempty"` // This is a standard url that can be redirected to.
}

// AlipayBrowserOrder represents an order creates for alipay inside
// browsers
type AlipayBrowserOrder struct {
	Order
	RedirectURL string `json:"redirectUrl"`
}

// AliPayNative is an order created inside a native app.
type AlipayNativeAppOrder struct {
	Order
	//FtcOrderID string `json:"ftcOrderId"` // Deprecate
	Param string `json:"param"`
}

// PaymentIntent contains the data describing user's intent to pay.
// The data are constructed prior to payment.
type PaymentIntent struct {
	product.Charge   // How much user should pay.
	product.Duration // How long the membership this payment purchased.

	SubsKind enum.OrderKind     `json:"subscriptionKind"`
	Wallet   Wallet             `json:"wallet"`
	Plan     product.IntentPlan `json:"plan"` // The plan to subscribe.
}
