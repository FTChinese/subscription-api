package wechat

import (
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

// WxResp contains the shared fields all wechat pay endpoints contains.
type WxResp struct {
	// return_code: SUCCESS/FAIL
	StatusCode string
	// return_msg
	StatusMessage string
	// appid
	AppID null.String
	// mch_id
	MID null.String
	// nonce_str
	Nonce null.String
	// sign
	Signature null.String
	// result_code: SUCCESS/FAIL
	ResultCode null.String
	// err_code
	ErrorCode null.String
	// err_code_des
	ErrorMessage null.String
}

func (r *WxResp) Populate(p wxpay.Params) {
	r.StatusCode = p.GetString("return_code")
	r.StatusMessage = p.GetString("return_msg")

	if v, ok := p["appid"]; ok {
		r.AppID = null.StringFrom(v)
	}

	if v, ok := p["mch_id"]; ok {
		r.MID = null.StringFrom(v)
	}

	if v, ok := p["nonce_str"]; ok {
		r.Nonce = null.StringFrom(v)
	}

	if v, ok := p["sign"]; ok {
		r.Signature = null.StringFrom(v)
	}

	if v, ok := p["result_code"]; ok {
		r.ResultCode = null.StringFrom(v)
	}

	if v, ok := p["err_code"]; ok {
		r.ErrorCode = null.StringFrom(v)
	}

	if v, ok := p["err_code_des"]; ok {
		r.ErrorMessage = null.StringFrom(v)
	}
}
