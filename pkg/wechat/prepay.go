package wechat

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

// UnifiedOrderResp is wechat's response for prepay.
type UnifiedOrder struct {
	BaseResp
	TradeType  null.String             `db:"trade_type"`
	PrepayID   null.String             `db:"prepay_id"`
	QRCode     null.String             `db:"qr_code"`
	MWebURL    null.String             `db:"mobile_redirect_url"`
	FtcOrderID string                  `db:"order_id"`
	Invalid    *render.ValidationError `db:"-"`
}

func (o UnifiedOrder) Params() wxpay.Params {
	p := o.BaseParams()

	p.SetString("prepay_id", o.PrepayID.String)

	return p
}

// NewUnifiedOrderResp creates converts PrePay from a wxpay.Params type.
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
func NewUnifiedOrderResp(orderID string, p wxpay.Params) UnifiedOrder {
	r := UnifiedOrder{
		FtcOrderID: orderID,
	}

	r.Populate(p)

	if v, ok := p["trade_type"]; ok {
		r.TradeType = null.StringFrom(v)
	}

	if v, ok := p["prepay_id"]; ok {
		r.PrepayID = null.StringFrom(v)
	}

	// For native pay.
	if v, ok := p["code_url"]; ok {
		r.QRCode = null.StringFrom(v)
	}

	if v, ok := p["mweb_url"]; ok {
		r.MWebURL = null.StringFrom(v)
	}

	return r
}
