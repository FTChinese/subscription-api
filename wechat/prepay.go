package wechat

import (
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

// UnifiedOrderResp is wechat's response for prepay.
type UnifiedOrderResp struct {
	Resp
	TradeType null.String
	PrepayID  null.String
	CodeURL   null.String
	MWebURL   null.String
}

func (o UnifiedOrderResp) Params() wxpay.Params {
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
func NewUnifiedOrderResp(p wxpay.Params) UnifiedOrderResp {
	r := UnifiedOrderResp{}

	r.Populate(p)

	if v, ok := p["trade_type"]; ok {
		r.TradeType = null.StringFrom(v)
	}

	if v, ok := p["prepay_id"]; ok {
		r.PrepayID = null.StringFrom(v)
	}

	// For native pay.
	if v, ok := p["code_url"]; ok {
		r.CodeURL = null.StringFrom(v)
	}

	if v, ok := p["mweb_url"]; ok {
		r.MWebURL = null.StringFrom(v)
	}

	return r
}
