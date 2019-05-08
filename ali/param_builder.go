package ali

import "github.com/smartwalle/alipay"

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
