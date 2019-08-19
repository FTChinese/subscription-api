package ali

import (
	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
)

// BrowserOrder represents an order creates for alipay inside
// browsers
type BrowserOrder struct {
	paywall.Subscription
	PayURL      string `json:"payUrl"` // Deprecate
	RedirectURL string `json:"redirectUrl"`
}

func NewBrowserOrder(subs paywall.Subscription, redirect string) BrowserOrder {
	return BrowserOrder{
		Subscription: subs,
		PayURL:       redirect,
		RedirectURL:  redirect,
	}
}

// AppOrder is an order created inside a native app.
type AppOrder struct {
	paywall.Subscription
	FtcOrderID string `json:"ftcOrderId"` // Deprecate
	Param      string `json:"param"`
}

func NewAppOrder(subs paywall.Subscription, query string) AppOrder {
	return AppOrder{
		Subscription: subs,
		FtcOrderID:   subs.ID,
		Param:        query,
	}
}

// BuildAppPay creates the data type required by alipay.TradeAppPay
func BuildAppPay(t alipay.TradePay) alipay.AliPayTradeAppPay {
	p := alipay.AliPayTradeAppPay{}
	p.NotifyURL = t.NotifyURL
	p.Subject = t.Subject
	p.OutTradeNo = t.OutTradeNo
	p.TotalAmount = t.TotalAmount
	p.ProductCode = ProductCodeApp.String()
	p.GoodsType = t.GoodsType

	return p
}

// BuildDesktopPay creates the data type required by
// alipay.TradePagePay.
func BuildDesktopPay(t alipay.TradePay) alipay.AliPayTradePagePay {
	p := alipay.AliPayTradePagePay{}
	p.NotifyURL = t.NotifyURL
	p.ReturnURL = t.ReturnURL
	p.Subject = t.Subject
	p.OutTradeNo = t.OutTradeNo
	p.TotalAmount = t.TotalAmount
	p.ProductCode = ProductCodeWeb.String()
	p.GoodsType = t.GoodsType

	return p
}

// BuildWapPay creates the data type required by
// alipay.TradeWapPay
func BuildWapPay(t alipay.TradePay) alipay.AliPayTradeWapPay {
	p := alipay.AliPayTradeWapPay{}
	p.NotifyURL = t.NotifyURL
	p.ReturnURL = t.ReturnURL
	p.Subject = t.Subject
	p.OutTradeNo = t.OutTradeNo
	p.TotalAmount = t.TotalAmount
	p.ProductCode = ProductCodeWeb.String()
	p.GoodsType = t.GoodsType

	return p
}
