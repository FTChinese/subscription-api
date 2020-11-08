package subs

import (
	"github.com/FTChinese/subscription-api/pkg/wechat"
)

// WxPayNativeAppIntent creates an order used by native apps.
type WxPayNativeAppIntent struct {
	Order            // Deprecated
	OrderField Order `json:"order"`
	//wechat.NativeAppParams // Deprecated
	Params wechat.NativeAppParams `json:"params"`
}

// WepayEmbedBrowserOrder responds to purchase made in wechat
// embedded browser.
// This is actually similar to AppOrder since they are all
// perform actions inside wechat app.
// It's a shame wechat cannot even use the same data structure
// for such insignificant differences.
type WxPayJSApiIntent struct {
	Order                         // Deprecated
	OrderField Order              `json:"order"`
	Params     wechat.JSApiParams `json:"params"`
}

// WxPayBrowserIntent creates order for payment via wechat
// made in browsers.
// For desktop browser, wechat send back a custom url
// for the client to generate a QR image;
// For mobile browser, wechat sends back a canonical url
// that can be redirected to.
// and MobileOrder into a single data structure.
type WxPayBrowserIntent struct {
	Order            // Deprecated
	OrderField Order `json:"order"`
	// TODO: rename json tag codeUrl to qrCode
	QRCode  string `json:"qrCodeUrl,omitempty"`         // Used by desktop browser. It is a custom url like wexin://wxpay/bizpayurl
	MWebURL string `json:"mobileRedirectUrl,omitempty"` // This is a standard url that can be redirected to.
}

// AlipayBrowserIntent represents an order creates for alipay inside
// browsers
type AlipayBrowserIntent struct {
	Order              // Deprecated
	OrderField  Order  `json:"order"`
	RedirectURL string `json:"redirectUrl"`
}

// AliPayNative is an order created inside a native app.
type AlipayNativeIntent struct {
	Order             // Deprecated
	OrderField Order  `json:"order"`
	Param      string `json:"param"`
}

type PaymentIntent struct {
	Checkout
	Order Order
}

// AliAppPayIntent build the data to be sent to native apps.
// The param is Alipsy sdk's signed string.
func (pi PaymentIntent) AliAppPayIntent(param string) AlipayNativeIntent {
	return AlipayNativeIntent{
		Order:      pi.Order,
		OrderField: pi.Order,
		Param:      param,
	}
}

// AliPayBrowserIntent build the data required to pay inside desktop or mobile browsers.
func (pi PaymentIntent) AliPayBrowserIntent(redirectURL string) AlipayBrowserIntent {
	return AlipayBrowserIntent{
		Order:       pi.Order,
		OrderField:  pi.Order,
		RedirectURL: redirectURL,
	}
}

func (pi PaymentIntent) WxPayDesktopIntent(wxOrder wechat.OrderResp) WxPayBrowserIntent {
	return WxPayBrowserIntent{
		Order:      pi.Order,
		OrderField: pi.Order,
		QRCode:     wxOrder.QRCode.String,
		MWebURL:    "",
	}
}

func (pi PaymentIntent) WxPayMobileIntent(wxOrder wechat.OrderResp) WxPayBrowserIntent {
	return WxPayBrowserIntent{
		Order:      pi.Order,
		OrderField: pi.Order,
		QRCode:     "",
		MWebURL:    wxOrder.MWebURL.String,
	}
}

func (pi PaymentIntent) WxPayJSApiIntent(p wechat.JSApiParams) WxPayJSApiIntent {
	return WxPayJSApiIntent{
		Order:      pi.Order,
		OrderField: pi.Order,
		Params:     p,
	}
}

func (pi PaymentIntent) WxNativeAppIntent(p wechat.NativeAppParams) WxPayNativeAppIntent {
	return WxPayNativeAppIntent{
		Order:      pi.Order,
		OrderField: pi.Order,
		Params:     p,
	}
}
