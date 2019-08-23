package wechat

import (
	"errors"
	"github.com/FTChinese/go-rest/view"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

// WxResp contains the shared fields all wechat pay endpoints contains.
type Resp struct {
	// return_code: SUCCESS/FAIL
	StatusCode string `db:"status_code"`
	// return_msg
	StatusMessage string `db:"status_message"`
	// appid
	AppID null.String `db:"app_id"`
	// mch_id
	MID null.String `db:"merchant_id"`
	// nonce_str
	Nonce null.String `db:"nonce"`
	// sign
	Signature null.String `db:"signature"`
	// result_code: SUCCESS/FAIL
	ResultCode null.String `db:"result_code"`
	// err_code
	ErrorCode null.String `db:"error_code"`
	// err_code_des
	ErrorMessage null.String `db:"error_message"`
}

// BaseParams turns Resp to a wxpay.Params.
// This is used to mock wechat pay's response.
// Used only for testing.
func (r Resp) BaseParams() wxpay.Params {
	p := make(wxpay.Params)

	p.SetString("return_code", r.StatusCode)
	p.SetString("return_msg", r.StatusMessage)
	p.SetString("appid", r.AppID.String)
	p.SetString("mch_id", r.MID.String)
	p.SetString("nonce_str", r.Nonce.String)
	p.SetString("result_code", r.ResultCode.String)

	return p
}

// Populate fills the fields of Resp from wxpay.Params
func (r *Resp) Populate(p wxpay.Params) {

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

// IsStatusValid checks if wechat reponse contains `return_code`.
// It does not check if `return_code` is SUCCESS of FAIL,
// following `wxpay` package's convention.
func (r Resp) IsStatusValid() error {

	switch r.StatusCode {
	case "":
		return errors.New("no return_code in XML")

	case wxpay.Fail, wxpay.Success:
		return nil

	default:
		return errors.New("return_code value is invalid in XML")
	}
}

// Validate wechat if wechat response is valid.
// This validation does not include notification since the
// `wxpay` package does not provide such capability.
// The `wxpay` pakcage already performs an initial check
// on `return_code` and `sign`.
// But its check is incomplete since `return_code == FAIL`
// is regarded as ok.
// You have to check if return_code == SUCCESS, appid, mch_id, result_code are valid.
func (r Resp) Validate(app PayApp) *view.Reason {
	if r.StatusCode == wxpay.Fail {
		reason := &view.Reason{
			Field: "status",
			Code:  "fail",
		}
		reason.SetMessage(r.StatusMessage)

		return reason
	}

	if r.ResultCode.String == wxpay.Fail {
		reason := &view.Reason{
			Field: "result",
			Code:  r.ErrorCode.String,
		}
		reason.SetMessage(r.ErrorMessage.String)

		return reason
	}

	if r.AppID.IsZero() {
		reason := &view.Reason{
			Field: "app_id",
			Code:  view.CodeInvalid,
		}
		reason.SetMessage("Missing app id")

		return reason
	}

	if r.MID.IsZero() {
		reason := &view.Reason{
			Field: "mch_id",
			Code:  view.CodeInvalid,
		}
		reason.SetMessage("Missing merchant id")

		return reason
	}

	if r.AppID.String != app.AppID {
		reason := &view.Reason{
			Field: "app_id",
			Code:  view.CodeInvalid,
		}
		reason.SetMessage("Missing or wrong app id")

		return reason
	}

	if r.MID.String != app.MchID {
		reason := &view.Reason{
			Field: "mch_id",
			Code:  view.CodeInvalid,
		}
		reason.SetMessage("Missing or wrong merchant id")

		return reason
	}

	return nil
}
