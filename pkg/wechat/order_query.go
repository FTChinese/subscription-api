package wechat

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
)

const (
	TradeStateSuccess  = "SUCCESS"
	TradeStateRefund   = "REFUND"
	TradeStateNotPay   = "NOTPAY"
	TradeStateClosed   = "CLOSED"
	TradeStateRevoked  = "REVOKED"
	TradeStatePaying   = "USERPAYING"
	TradeStatePayError = "PAYERROR"
)

// OrderQueryResp for https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
type OrderQueryResp struct {
	Resp
	// 以下字段在return_code 和result_code都为SUCCESS的时候有返回
	OpenID    null.String `db:"open_id"`
	TradeType null.String `db:"trade_type"` // APP
	// SUCCESS—支付成功
	// REFUND—转入退款
	// NOTPAY—未支付
	// CLOSED—已关闭
	// REVOKED—已撤销（刷卡支付）
	// USERPAYING--用户支付中
	// PAYERROR--支付失败(其他原因，如银行返回失败)
	// This is `trade_state` field in wechat response.
	TradeState null.String `db:"trade_state"`
	// This is the `trade_state_desc` field.
	TradeStateDesc null.String `db:"trade_state_desc"`
	BankType       null.String `db:"bank_type"`
	TotalFee       null.Int    `db:"total_fee"`
	Currency       null.String `db:"currency"`
	TransactionID  null.String `db:"transaction_id"`
	FTCOrderID     null.String `db:"ftc_order_id"`
	TimeEnd        null.String `db:"time_end"` // 20141030133525
}

func NewOrderQueryResp(p wxpay.Params) OrderQueryResp {
	r := OrderQueryResp{}

	r.Populate(p)

	if v, ok := p["openid"]; ok {
		r.OpenID = null.StringFrom(v)
	}

	if v, ok := p["trade_type"]; ok {
		r.TradeType = null.StringFrom(v)
	}

	if v, ok := p["trade_state"]; ok {
		r.TradeState = null.StringFrom(v)
	}

	if v, ok := p["bank_type"]; ok {
		r.BankType = null.StringFrom(v)
	}

	if price := p.GetInt64("total_fee"); price != 0 {
		r.TotalFee = null.IntFrom(price)
	}

	if v, ok := p["fee_type"]; ok {
		r.Currency = null.StringFrom(v)
	}

	if v, ok := p["transaction_id"]; ok {
		r.TransactionID = null.StringFrom(v)
	}

	if v, ok := p["out_trade_no"]; ok {
		r.FTCOrderID = null.StringFrom(v)
	}

	if v, ok := p["time_end"]; ok {
		r.TimeEnd = null.StringFrom(v)
	}

	if v, ok := p["trade_state_desc"]; ok {
		r.TradeStateDesc = null.StringFrom(v)
	}

	return r
}

// ValidateTradeState ensures the transaction is actually done successfully.
func (r OrderQueryResp) ValidateTradeState() *render.ValidationError {
	if r.TradeState.String == TradeStateSuccess {
		return nil
	}

	return &render.ValidationError{
		Message: r.TradeStateDesc.String,
		Field:   "trade_state",
		Code:    render.InvalidCode(r.TradeState.String),
	}
}
