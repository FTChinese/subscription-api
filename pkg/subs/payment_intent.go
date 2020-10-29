package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/objcoding/wxpay"
	"github.com/smartwalle/alipay"
)

// WxpayNativeAppIntent creates an order used by native apps.
type WxpayNativeAppIntent struct {
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
type WxpayEmbedBrowserIntent struct {
	Order
	Params wechat.InWxBrowserParams `json:"params"`
}

// WxpayBrowserIntent creates order for payment via wechat
// made in browsers.
// For desktop browser, wechat send back a custom url
// for the client to generate a QR image;
// For mobile browser, wechat sends back a canonical url
// that can be redirected to.
// and MobileOrder into a single data structure.
type WxpayBrowserIntent struct {
	Order
	// TODO: rename json tag codeUrl to qrCode
	QRCode  string `json:"qrCodeUrl,omitempty"`         // Used by desktop browser. It is a custom url like wexin://wxpay/bizpayurl
	MWebURL string `json:"mobileRedirectUrl,omitempty"` // This is a standard url that can be redirected to.
}

// AlipayBrowserIntent represents an order creates for alipay inside
// browsers
type AlipayBrowserIntent struct {
	Order
	RedirectURL string `json:"redirectUrl"`
}

// AliPayNative is an order created inside a native app.
type AlipayNativeIntent struct {
	Order
	Param string `json:"param"`
}

type PaymentIntent struct {
	Checkout
	Order      Order
	WebhookURL string
}

func (pi PaymentIntent) ProratedOrders() []ProratedOrder {
	if pi.Kind != enum.OrderKindUpgrade {
		return nil
	}
	if pi.Wallet.Sources == nil || len(pi.Wallet.Sources) == 0 {
		return nil
	}

	now := chrono.TimeNow()

	for i, v := range pi.Wallet.Sources {
		v.UpgradeOrderID = pi.Order.ID
		if pi.IsFree {
			v.ConsumedUTC = now
		}

		pi.Wallet.Sources[i] = v
	}

	return pi.Wallet.Sources
}

// AliAppPayParam builds the data that can be to tell Ali what product user is buying.
// The return data has to be signed by Alipay sdk
// before sending it to client.
func (pi PaymentIntent) AliAppPayParam() alipay.AliPayTradeAppPay {
	return alipay.AliPayTradeAppPay{
		TradePay: alipay.TradePay{
			NotifyURL:   pi.WebhookURL,
			Subject:     pi.Item.Plan.PaymentTitle(pi.Kind),
			OutTradeNo:  pi.Order.ID,
			TotalAmount: pi.Payable.AliPrice(),
			ProductCode: ali.ProductCodeApp.String(),
			GoodsType:   "0",
		},
	}
}

// AliAppPayIntent build the data to be sent to native apps.
// The param is Alipsy sdk's signed string.
func (pi PaymentIntent) AliAppPayIntent(param string) AlipayNativeIntent {
	return AlipayNativeIntent{
		Order: pi.Order,
		Param: param,
	}
}

func (pi PaymentIntent) AliDesktopPayParam(retURL string) alipay.AliPayTradePagePay {
	return alipay.AliPayTradePagePay{
		TradePay: alipay.TradePay{
			NotifyURL:   pi.WebhookURL,
			ReturnURL:   retURL,
			Subject:     pi.Item.Plan.PaymentTitle(pi.Order.Kind),
			OutTradeNo:  pi.Order.ID,
			TotalAmount: pi.Payable.AliPrice(),
			ProductCode: ali.ProductCodeWeb.String(),
			GoodsType:   "0",
		},
	}
}

func (pi PaymentIntent) AliWapPayParam(retURL string) alipay.AliPayTradeWapPay {
	return alipay.AliPayTradeWapPay{
		TradePay: alipay.TradePay{
			NotifyURL:   pi.WebhookURL,
			ReturnURL:   retURL,
			Subject:     pi.Item.Plan.PaymentTitle(pi.Order.Kind),
			OutTradeNo:  pi.Order.ID,
			TotalAmount: pi.Payable.AliPrice(),
			ProductCode: ali.ProductCodeWeb.String(),
			GoodsType:   "0",
		},
	}
}

func (pi PaymentIntent) AliPayBrowserIntent(redirectURL string) AlipayBrowserIntent {
	return AlipayBrowserIntent{
		Order:       pi.Order,
		RedirectURL: redirectURL,
	}
}
func (pi PaymentIntent) WxPayParam(wxParam wechat.UnifiedOrderConfig) wxpay.Params {

	p := make(wxpay.Params)
	p.SetString("body", pi.Item.Plan.PaymentTitle(pi.Kind))
	p.SetString("out_trade_no", pi.Order.ID)
	p.SetInt64("total_fee", pi.Order.AmountInCent())
	p.SetString("spbill_create_ip", wxParam.IP)
	p.SetString("notify_url", pi.WebhookURL)
	// APP for native app
	// NATIVE for web site
	// JSAPI for web page opend inside wechat browser
	p.SetString("trade_type", wxParam.TradeType.String())

	switch wxParam.TradeType {
	case wechat.TradeTypeDesktop:
		p.SetString("product_id", pi.Item.Plan.ID)

	case wechat.TradeTypeJSAPI:
		p.SetString("openid", wxParam.OpenID)
	}

	return p
}
