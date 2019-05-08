package wechat

import (
	"github.com/objcoding/wxpay"
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
