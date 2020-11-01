package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/smartwalle/alipay"
	"strconv"
	"time"
)

func priceToCent(s string) (int64, error) {
	if s == "" {
		return -1, nil
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return int64(f * 100), nil
}

func mustPriceToCent(s string) int64 {
	i, err := priceToCent(s)
	if err != nil {
		panic(err)
	}

	return i
}

// PaymentResult unifies ali and wx webhook payload, or query order.
// TODO: add payment method?
type PaymentResult struct {
	// For Alipay `trade_status` field:
	// WAIT_BUYER_PAY（交易创建，等待买家付款）
	// TRADE_CLOSED（未付款交易超时关闭，或支付完成后全额退款）
	// TRADE_SUCCESS（交易支付成功）
	// TRADE_FINISHED（交易结束，不可退款）
	// For Wechat `trade_state` field:
	// SUCCESS—支付成功
	// REFUND—转入退款
	// NOTPAY—未支付
	// CLOSED—已关闭
	// REVOKED—已撤销（刷卡支付）
	// USERPAYING--用户支付中
	// PAYERROR--支付失败(其他原因，如银行返回失败)
	PaymentState string `json:"paymentState"`
	// For wechat `trade_state_desc` field.
	PaymentStateDesc string `json:"paymentStateDesc"`
	// In cent.
	// For alipay, we use the total_amount, which is
	// 交易的订单金额，单位为元，两位小数。该参数的值为支付时传入的total_amount.
	// For our purpose, that is the amount we actually charged user.
	Amount        int64       `json:"totalFee"` // In cent
	TransactionID string      `json:"transactionId"`
	OrderID       string      `json:"ftcOrderId"`
	ConfirmedAt   chrono.Time `json:"paidAt"`
}

// IsOrderPaid checks if a wx or ali order is paid.
// Do not call it for webhook since the payload does not contain these fields.
func (r PaymentResult) IsOrderPaid() bool {
	switch r.PaymentState {
	case ali.TradeStatusSuccess,
		ali.TradeStatusFinished,
		wechat.TradeStateSuccess:
		return true
	}

	return false
}

func (r PaymentResult) ConfirmError(err error, retry bool) *ConfirmError {
	return &ConfirmError{
		OrderID: r.OrderID,
		Message: err.Error(),
		Retry:   retry,
	}
}

// NewWxWebhookResult builds PaymentResult from wechat pay webhook notification.
func NewWxWebhookResult(payload wechat.Notification) PaymentResult {
	return PaymentResult{
		PaymentState:     "",
		PaymentStateDesc: "",
		Amount:           payload.TotalFee.Int64,
		TransactionID:    payload.TransactionID.String,
		OrderID:          payload.FTCOrderID.String,
		ConfirmedAt:      chrono.TimeFrom(dt.MustParseWxTime(payload.TimeEnd.String)),
	}
}

// NewWxPayResult creates a new PaymentResult from the result of querying wechat order.
// The payment status should be returned to client as is, whether it is paid or not.
func NewWxPayResult(r wechat.OrderQueryResp) PaymentResult {
	// If the payment is not done, we should not add confirmation time.
	var confirmedAt time.Time
	if r.TimeEnd.Valid {
		confirmedAt = dt.MustParseWxTime(r.TimeEnd.String)
	}
	return PaymentResult{
		PaymentState:     r.TradeState.String,
		PaymentStateDesc: r.TradeStateDesc.String,
		Amount:           r.TotalFee.Int64,
		TransactionID:    r.TransactionID.String,
		OrderID:          r.FTCOrderID.String,
		ConfirmedAt:      chrono.TimeFrom(confirmedAt),
	}
}

func NewAliPayResult(r *alipay.AliPayTradeQueryResponse) PaymentResult {
	return PaymentResult{
		PaymentState:     r.AliPayTradeQuery.TradeStatus,
		PaymentStateDesc: ali.TradeStatusMsg[r.AliPayTradeQuery.TradeStatus],
		Amount:           mustPriceToCent(r.AliPayTradeQuery.TotalAmount),
		TransactionID:    r.AliPayTradeQuery.TradeNo,
		OrderID:          r.AliPayTradeQuery.OutTradeNo,
		ConfirmedAt:      chrono.TimeNow(),
	}
}

// NewAliWebhookResult builds PaymentResult from alipay webhook notification.
func NewAliWebhookResult(payload *alipay.TradeNotification) (PaymentResult, error) {
	a, err := priceToCent(payload.TotalAmount)
	if err != nil {
		return PaymentResult{}, err
	}

	return PaymentResult{
		PaymentState:     payload.TradeStatus,
		PaymentStateDesc: ali.TradeStatusMsg[payload.TradeStatus],
		Amount:           a,
		TransactionID:    payload.TradeNo,
		OrderID:          payload.OutTradeNo,
		ConfirmedAt:      chrono.TimeFrom(dt.MustParseAliTime(payload.GmtPayment)),
	}, nil
}
