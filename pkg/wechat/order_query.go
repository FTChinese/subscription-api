package wechat

import (
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

// OrderQueryParams contains parameters required to search an order.
type OrderQueryParams struct {
	AppID      string `map:"appid"`          // 微信支付分配的公众账号ID
	MchID      string `map:"mch_id"`         // 微信支付分配的商户号
	TxnID      string `map:"transaction_id"` // 微信的订单号，建议优先使用
	OutTradeNo string `map:"out_trade_no"`   // 商户系统内部订单号
	Nonce      string `map:"nonce_str"`      // 随机字符串，不长于32位
	Sign       string `map:"sign"`           // 通过签名算法计算得出的签名值
	SignType   string `map:"sign_type"`      // 签名类型，目前支持HMAC-SHA256和MD5，默认为MD5
}

func NewOrderQueryParams(orderID string) OrderQueryParams {
	return OrderQueryParams{
		AppID:      "",
		MchID:      "",
		TxnID:      "",
		OutTradeNo: orderID,
		Nonce:      "",
		Sign:       "",
		SignType:   "",
	}
}

func (q OrderQueryParams) Marshal() wxpay.Params {
	return make(wxpay.Params).
		SetString(keyOrderID, q.OutTradeNo)
}

// OrderQueryResp for https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
type OrderQueryResp struct {
	FTCOrderID    string `map:"out_trade_no"`
	TimeEnd       string `map:"time_end"` // 订单支付时间，格式为yyyyMMddHHmmss，如2009年12月25日9点10分10秒表示为20091225091010
	TotalFee      int64  `map:"total_fee"`
	TransactionID string `map:"transaction_id"`
	// SUCCESS—支付成功
	// REFUND—转入退款
	// NOTPAY—未支付
	// CLOSED—已关闭
	// REVOKED—已撤销（刷卡支付）
	// USERPAYING--用户支付中
	// PAYERROR--支付失败(其他原因，如银行返回失败)
	TradeState     string `map:"trade_state"`
	TradeStateDesc string `map:"trade_state_desc"` //对当前查询订单状态的描述和下一步操作的指引
}

func NewOrderQueryResp(p wxpay.Params) OrderQueryResp {
	return OrderQueryResp{
		FTCOrderID:     GetOrderID(p),
		TimeEnd:        p.GetString(keyEndTime),
		TotalFee:       p.GetInt64(keyTotalAmount),
		TransactionID:  p.GetString(keyTxnID),
		TradeState:     p.GetString("trade_state"),
		TradeStateDesc: p.GetString("trade_state_desc"),
	}
}
