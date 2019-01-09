package wechat

import (
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

// PrePay contains the response data from Wechat unified order.
type PrePay struct {
	StatusCode    string
	StatusMessage string
	AppID         null.String
	MID           null.String
	Nonce         null.String
	Signature     null.String
	IsSuccess     bool
	ResultCode    null.String
	ResultMessage null.String
	TradeType     null.String
	PrePayID      null.String
}

// NewPrePay creates converts PrePay from a wxpay.Params type.
func NewPrePay(r wxpay.Params) PrePay {
	p := PrePay{
		StatusCode:    r.GetString("return_code"),
		StatusMessage: r.GetString("return_msg"),
		IsSuccess:     r.GetString("result_code") == "SUCCESS",
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

	if v, ok := r["err_code"]; ok {
		p.ResultCode = null.StringFrom(v)
	}
	if v, ok := r["err_code_des"]; ok {
		p.ResultMessage = null.StringFrom(v)
	}

	if v, ok := r["trade_type"]; ok {
		p.TradeType = null.StringFrom(v)
	}

	if v, ok := r["prepay_id"]; ok {
		p.PrePayID = null.StringFrom(v)
	}

	return p
}

// SavePrePay saves Wechat prepay response for future analysis.
func (env Env) SavePrePay(p PrePay) error {
	query := `
	INSERT INTO premium.log_wx_prepay
	SET status_code = ?,
		status_message = ?,
		app_id = ?,
		merchant_id = ?,
		nonce = ?,
		signature = ?,
		is_success = ?,
		result_code = ?,
		result_message = ?,
		trade_type = ?,
		prepay_id = ?`

	_, err := env.DB.Exec(query,
		p.StatusCode,
		p.StatusMessage,
		p.AppID,
		p.MID,
		p.Nonce,
		p.Signature,
		p.IsSuccess,
		p.ResultCode,
		p.ResultMessage,
		p.TradeType,
		p.PrePayID,
	)

	if err != nil {
		return err
	}

	return nil
}
