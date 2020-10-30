package wechat

import (
	"errors"
	"fmt"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

const (
	ResultCodeOrderNotFound = "ORDERNOTEXIST"
	ResultCodeSystemError   = "SYSTEMERROR"
)

// BaseResp contains the shared fields all wechat pay endpoints return.
type BaseResp struct {
	// return_code: SUCCESS/FAIL.
	// 此字段是通信标识，非交易标识，交易是否成功需要查看trade_state来判断
	ReturnCode string `db:"status_code"`
	// return_msg. 返回信息，如非空，为错误原因
	ReturnMessage string `db:"status_message"`
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

func NewBaseResp(p wxpay.Params) BaseResp {
	r := BaseResp{
		ReturnCode:    p.GetString("return_code"),
		ReturnMessage: p.GetString("return_msg"),
		AppID:         null.String{},
		MID:           null.String{},
		Nonce:         null.String{},
		Signature:     null.String{},
		ResultCode:    null.String{},
		ErrorCode:     null.String{},
		ErrorMessage:  null.String{},
	}

	v, ok := p["appid"]
	r.AppID = null.NewString(v, ok)

	v, ok = p["mch_id"]
	r.MID = null.NewString(v, ok)

	v, ok = p["nonce_str"]
	r.Nonce = null.NewString(v, ok)

	v, ok = p["sign"]
	r.Signature = null.NewString(v, ok)

	v, ok = p["result_code"]
	r.ResultCode = null.NewString(v, ok)

	v, ok = p["err_code"]
	r.ErrorCode = null.NewString(v, ok)

	v, ok = p["err_code_des"]
	r.ErrorMessage = null.NewString(v, ok)

	return r
}

// Populate fills the fields of BaseResp from wxpay.Params
func (r *BaseResp) Populate(p wxpay.Params) {

	r.ReturnCode = p.GetString("return_code")
	r.ReturnMessage = p.GetString("return_msg")

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

// ValidateResponded checks if wechat send back a response by checking if it contains `return_code`.
// It does not check if `return_code` is SUCCESS of FAIL,
// following `wxpay` package's convention.
func (r BaseResp) ValidateResponded() error {

	switch r.ReturnCode {
	case "":
		return errors.New("no return_code in XML")

	case wxpay.Fail, wxpay.Success:
		return nil

	default:
		return fmt.Errorf("invalid reponse: %s, %s", r.ReturnCode, r.ReturnMessage)
	}
}

func (r BaseResp) ValidateReturnSuccess() error {
	if r.ReturnCode == wxpay.Success {
		return nil
	}

	return fmt.Errorf("wxpay api return_code not success: %s, %s", r.ReturnCode, r.ReturnCode)
}

func (r BaseResp) ValidateResultSuccess() error {
	if r.ResultCode.Valid && r.ResultCode.String == wxpay.Success {
		return nil
	}

	return fmt.Errorf("wxpay api result_code not success: %s, %s", r.ErrorCode.String, r.ErrorMessage.String)
}

func (r BaseResp) ValidateIdentity(app PayApp) error {
	if r.AppID.String != app.AppID {
		return errors.New("wxpay appid mismatched")
	}

	if r.MID.String != app.MchID {
		return errors.New("wxpay mch_id mismatched")
	}

	return nil
}

func (r BaseResp) IsSuccess() bool {
	return r.ReturnCode == wxpay.Success
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
func (r BaseResp) Validate(app PayApp) error {

	if err := r.ValidateReturnSuccess(); err != nil {
		return err
	}

	if err := r.ValidateResultSuccess(); err != nil {
		return err
	}

	if err := r.ValidateIdentity(app); err != nil {
		return err
	}

	return nil
}
