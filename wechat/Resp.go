package wechat

import (
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

// WxResp contains the shared fields all wechat pay endpoints contains.
type WxResp struct {
	StatusCode    string // SUCCESS/FAIL
	StatusMessage string
	AppID         null.String
	MID           null.String
	Nonce         null.String
	Signature     null.String
	ResultCode    null.String // SUCCESS/FAIL
	ErrorCode     null.String
	ErrorMessage  null.String
}

func (r *WxResp) Populate(p wxpay.Params)  {
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
