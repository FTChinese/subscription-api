package wechat

import (
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

// UnifiedOrderResp contains the response data from Wechat unified order.
type UnifiedOrderResp struct {
	StatusCode       string
	StatusMessage    string
	AppID            null.String
	MID              null.String
	Nonce            null.String
	Signature        null.String
	ResultCode       null.String
	ErrorCode        null.String
	ErrorDescription null.String
	TradeType        null.String
	PrepayID         null.String
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
// appid:wxacddf1c20516eb69
// mch_id:1504993271
// nonce_str:aOyCOfOvWZQZkRwp
// ]
func NewUnifiedOrderResp(r wxpay.Params) UnifiedOrderResp {
	p := UnifiedOrderResp{
		StatusCode:    r.GetString("return_code"),
		StatusMessage: r.GetString("return_msg"),
	}

	if v, ok := r["appid"]; ok {
		p.AppID = null.StringFrom(v)
	}

	if v, ok := r["mch_id"]; ok {
		p.MID = null.StringFrom(v)
	}

	if v, ok := r["nonce_str"]; ok {
		p.Nonce = null.StringFrom(v)
	}

	if v, ok := r["sign"]; ok {
		p.Signature = null.StringFrom(v)
	}

	if v, ok := r["result_code"]; ok {
		p.ResultCode = null.StringFrom(v)
	}

	if v, ok := r["err_code"]; ok {
		p.ErrorCode = null.StringFrom(v)
	}

	if v, ok := r["err_code_des"]; ok {
		p.ErrorDescription = null.StringFrom(v)
	}

	if v, ok := r["trade_type"]; ok {
		p.TradeType = null.StringFrom(v)
	}

	if v, ok := r["prepay_id"]; ok {
		p.PrepayID = null.StringFrom(v)
	}

	return p
}
