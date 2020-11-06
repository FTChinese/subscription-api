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
	ReturnMessage null.String `db:"status_message"`
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
		ReturnMessage: null.String{},
		AppID:         null.String{},
		MID:           null.String{},
		Nonce:         null.String{},
		Signature:     null.String{},
		ResultCode:    null.String{},
		ErrorCode:     null.String{},
		ErrorMessage:  null.String{},
	}

	v := p.GetString("return_msg")
	r.ReturnMessage = null.NewString(v, v == "")

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

// IsBadRequest tests if the `return_code` field is SUCCESS. Treat FAIL as 400 Bad Request.
func (r BaseResp) IsBadRequest() bool {
	return r.ReturnCode != wxpay.Success
}

func (r BaseResp) BadRequestMsg() string {
	return fmt.Sprintf("wxpay api bad requeest: %s, %s", r.ReturnCode, r.ReturnMessage.String)
}

// IsUnprocessable tests if `result_code` field is SUCCESS. Treat FAIL as 422 Unprocessable.
func (r BaseResp) IsUnprocessable() bool {
	return r.ResultCode.IsZero() || r.ResultCode.String != wxpay.Success
}

func (r BaseResp) UnprocessableMsg() string {
	return fmt.Sprintf("wxpay api unprocessable: %s - %s - %s", r.ResultCode.String, r.ErrorCode.String, r.ErrorMessage.String)
}

func (r BaseResp) ValidateIdentity(app PayApp) error {
	if r.AppID.String != app.AppID {
		return errors.New("wxpay api response: appid mismatched")
	}

	if r.MID.String != app.MchID {
		return errors.New("wxpay api response: mch_id mismatched")
	}

	return nil
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

	if r.IsBadRequest() {
		return fmt.Errorf("%s", r.BadRequestMsg())
	}

	if r.IsUnprocessable() {
		return fmt.Errorf("%s", r.UnprocessableMsg())
	}

	if err := r.ValidateIdentity(app); err != nil {
		return err
	}

	return nil
}
