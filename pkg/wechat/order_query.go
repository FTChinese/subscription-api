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
	BaseResp
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
	Invalid        *render.ValidationError
}

func NewOrderQueryResp(p wxpay.Params) OrderQueryResp {
	r := OrderQueryResp{
		BaseResp: NewBaseResp(p),
	}

	//r.Populate(p)
	v, ok := p["openid"]
	r.OpenID = null.NewString(v, ok)

	v, ok = p["trade_type"]
	r.TradeType = null.NewString(v, ok)

	v, ok = p["trade_state"]
	r.TradeState = null.NewString(v, ok)

	v, ok = p["bank_type"]
	r.BankType = null.NewString(v, ok)

	price := p.GetInt64("total_fee")
	r.TotalFee = null.NewInt(price, price != 0)

	v, ok = p["fee_type"]
	r.Currency = null.NewString(v, ok)

	v, ok = p["transaction_id"]
	r.TransactionID = null.NewString(v, ok)

	v, ok = p["out_trade_no"]
	r.FTCOrderID = null.NewString(v, ok)

	// 订单支付时间，格式为yyyyMMddHHmmss，如2009年12月25日9点10分10秒表示为20091225091010
	v, ok = p["time_end"]
	r.TimeEnd = null.NewString(v, ok)

	// 对当前查询订单状态的描述和下一步操作的指引
	v, ok = p["trade_state_desc"]
	r.TradeStateDesc = null.NewString(v, ok)

	return r
}
