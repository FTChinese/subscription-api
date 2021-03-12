package subs

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/wechat"
)

// PaymentTitle is used as the value of `subject` for alipay,
// and `body` for wechat pay.
// * 订阅FT中文网标准会员/年
// * 订阅FT中文网高端会员/年
// * 购买FT中文网标准会员/年
// * 购买FT中文网高端会员/年
func PaymentTitle(k enum.OrderKind, e price.Edition) string {
	var prefix string

	switch k {
	case enum.OrderKindCreate:
	case enum.OrderKindRenew:
	case enum.OrderKindUpgrade:
		prefix = "订阅"

	case enum.OrderKindAddOn:
		prefix = "购买"

	default:
	}

	return fmt.Sprintf("%sFT中文网%s", prefix, e.StringCN())
}

// WxPayNativeAppIntent creates an order used by native apps.
type WxPayNativeAppIntent struct {
	Order  Order                  `json:"order"`
	Params wechat.NativeAppParams `json:"params"`
}

func NewWxNativeAppIntent(order Order, p wechat.NativeAppParams) WxPayNativeAppIntent {
	return WxPayNativeAppIntent{
		Order:  order,
		Params: p,
	}
}

// WepayEmbedBrowserOrder responds to purchase made in wechat
// embedded browser.
// This is actually similar to AppOrder since they are all
// perform actions inside wechat app.
// It's a shame wechat cannot even use the same data structure
// for such insignificant differences.
type WxPayJSApiIntent struct {
	Order  Order              `json:"order"`
	Params wechat.JSApiParams `json:"params"`
}

func NewWxPayJSApiIntent(order Order, p wechat.JSApiParams) WxPayJSApiIntent {
	return WxPayJSApiIntent{
		Order:  order,
		Params: p,
	}
}

// WxPayBrowserIntent creates order for payment via wechat
// made in browsers.
// For desktop browser, wechat send back a custom url
// for the client to generate a QR image;
// For mobile browser, wechat sends back a canonical url
// that can be redirected to.
// and MobileOrder into a single data structure.
type WxPayBrowserIntent struct {
	Order   Order  `json:"order"`
	QRCode  string `json:"qrCodeUrl,omitempty"`         // Used by desktop browser. It is a custom url like wexin://wxpay/bizpayurl
	MWebURL string `json:"mobileRedirectUrl,omitempty"` // This is a standard url that can be redirected to.
}

func NewWxPayDesktopIntent(order Order, wxOrder wechat.OrderResp) WxPayBrowserIntent {
	return WxPayBrowserIntent{
		Order:   order,
		QRCode:  wxOrder.QRCode.String,
		MWebURL: "",
	}
}

// AlipayBrowserIntent represents an order creates for alipay inside
// browsers
type AlipayBrowserIntent struct {
	Order       Order  `json:"order"`
	RedirectURL string `json:"redirectUrl"`
}

func NewAliPayBrowserIntent(order Order, redirectURL string) AlipayBrowserIntent {
	return AlipayBrowserIntent{
		Order:       order,
		RedirectURL: redirectURL,
	}
}

func NewWxPayMobileIntent(order Order, wxOrder wechat.OrderResp) WxPayBrowserIntent {
	return WxPayBrowserIntent{
		Order:   order,
		QRCode:  "",
		MWebURL: wxOrder.MWebURL.String,
	}
}

// AliPayNative is an order created inside a native app.
type AlipayNativeIntent struct {
	Order Order  `json:"order"`
	Param string `json:"param"`
}

func NewAliAppPayIntent(order Order, param string) AlipayNativeIntent {
	return AlipayNativeIntent{
		Order: order,
		Param: param,
	}
}
