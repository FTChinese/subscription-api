package wechat

import (
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/go-rest/view"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

const (
	ResultCodeOrderNotFound = "ORDERNOTEXIST"
	ResultCodeSystemError   = "SYSTEMERROR"
)

// WxResp contains the shared fields all wechat pay endpoints contains.
type Resp struct {
	// return_code: SUCCESS/FAIL.
	// 此字段是通信标识，非交易标识，交易是否成功需要查看trade_state来判断
	StatusCode string `db:"status_code"`
	// return_msg. 返回信息，如非空，为错误原因
	StatusMessage string `db:"status_message"`
	// 以下字段在return_code为SUCCESS的时候有返回
	// appid
	AppID null.String `db:"app_id"`
	// mch_id
	MID null.String `db:"merchant_id"`
	// nonce_str
	Nonce null.String `db:"nonce"`
	// sign
	Signature null.String `db:"signature"`
	// result_code: SUCCESS/FAIL
	// 业务结果
	ResultCode null.String `db:"result_code"`
	// err_code 错误代码
	// ORDERNOTEXIST
	// SYSTEMERROR
	ErrorCode null.String `db:"error_code"`
	// err_code_des
	// 错误代码描述
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

// Validate wechat if wechat response itself is valid.
// It does not indicate whether the payment is success or not.
// This validation does not include notification since the
// `wxpay` package does not provide such capability.
// The `wxpay` pakcage already performs an initial check
// on `return_code` and `sign`.
// But its check is incomplete since `return_code == FAIL`
// is regarded as ok.
// You have to check if return_code == SUCCESS, appid, mch_id, result_code are valid.
func (r Resp) Validate(app PayApp) *render.ValidationError {
	// If `return_code` is FAIL
	// 此字段是通信标识，非交易标识，交易是否成功需要查看trade_state来判断
	if r.StatusCode == wxpay.Fail {
		reason := &render.ValidationError{
			Message: r.StatusCode + ":" + r.StatusMessage,
			Field:   "return_code",
			Code:    render.CodeInvalid,
		}

		return reason
	}

	// If `result_code` is FAIL
	if r.ResultCode.String == wxpay.Fail {
		reason := &render.ValidationError{
			Message: r.ErrorCode.String + ":" + r.ErrorMessage.String,
			Field:   "result_code",
			Code:    render.CodeInvalid,
		}

		return reason
	}

	if r.AppID.IsZero() {
		reason := &render.ValidationError{
			Message: "Missing app id",
			Field:   "app_id",
			Code:    view.CodeInvalid,
		}

		return reason
	}

	if r.MID.IsZero() {
		reason := &render.ValidationError{
			Message: "Missing merchant id",
			Field:   "mch_id",
			Code:    view.CodeInvalid,
		}

		return reason
	}

	if r.AppID.String != app.AppID {
		reason := &render.ValidationError{
			Message: "Missing or wrong app id",
			Field:   "app_id",
			Code:    view.CodeInvalid,
		}

		return reason
	}

	if r.MID.String != app.MchID {
		reason := &render.ValidationError{
			Message: "Missing or wrong merchant id",
			Field:   "mch_id",
			Code:    view.CodeInvalid,
		}

		return reason
	}

	return nil
}

func (r Resp) IsOrderNotFound() bool {
	return r.ResultCode.String == ResultCodeOrderNotFound
}
