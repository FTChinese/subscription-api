package wepay

import (
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

// PrepayOrder is the response send to client
type PrepayOrder struct {
	FtcOrderID string  `json:"ftcOrderId"`
	Price      float64 `json:"price"`
	AppID      string  `json:"appid"`
	PartnerID  string  `json:"partnerid"`
	PrepayID   string  `json:"prepayid"`
	Package    string  `json:"package"`
	Nonce      string  `json:"noncestr"`
	Timestamp  string  `json:"timestamp"`
	Signature  string  `json:"sign"`
}

// PrepayResp contains the response data from Wechat unified order.
type PrepayResp struct {
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

// NewPrepayResp creates converts PrePay from a wxpay.Params type.
func NewPrepayResp(r wxpay.Params) PrepayResp {
	p := PrepayResp{
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

// SavePrepayResp saves Wechat prepay response for future analysis.
func (env Env) SavePrepayResp(orderID string, p PrepayResp) error {
	query := `
	INSERT INTO premium.log_wx_prepay
	SET order_id = ?,
		status_code = ?,
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
		orderID,
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
