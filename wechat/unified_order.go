package wechat

import (
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

const wxNotifyURL = "http://www.ftacademy.cn/api/v1/callback/wxpay"

// GenerateUnifiedOrder to be used to request for prepay id.
func GenerateUnifiedOrder(plan paywall.Plan, userIP, orderID string) wxpay.Params {

	p := make(wxpay.Params)
	p.SetString("body", plan.Description)
	p.SetString("out_trade_no", orderID)
	p.SetInt64("total_fee", plan.PriceForWx())
	p.SetString("spbill_create_ip", userIP)
	p.SetString("notify_url", wxNotifyURL)
	p.SetString("trade_type", "APP")

	return p
}

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
