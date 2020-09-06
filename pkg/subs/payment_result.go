package subs

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/smartwalle/alipay"
	"strconv"
)

func priceToCent(s string) (int64, error) {
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

// PaymentResult unifies ali and wx webhook notification.
type PaymentResult struct {
	// Alipay:
	// WAIT_BUYER_PAY（交易创建，等待买家付款）
	// TRADE_CLOSED（未付款交易超时关闭，或支付完成后全额退款）
	// TRADE_SUCCESS（交易支付成功）
	// TRADE_FINISHED（交易结束，不可退款）
	// Wechat:
	// SUCCESS, REFUND, NOTPAY, CLOSED, REVOKED, USERPAYING, PAYERROR
	PaymentState     string `json:"paymentState"`
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

// NewWxQueryResult creates a new PaymentResult from the result of querying wechat order.
func NewWxQueryResult(r wechat.OrderQueryResp) PaymentResult {
	return PaymentResult{
		PaymentState:     r.TradeState.String,
		PaymentStateDesc: r.TradeStateDesc.String,
		Amount:           r.TotalFee.Int64,
		TransactionID:    r.TransactionID.String,
		OrderID:          r.FTCOrderID.String,
		ConfirmedAt:      chrono.TimeFrom(dt.MustParseWxTime(r.TimeEnd.String)),
	}
}

func NewAliQueryResult(r *alipay.AliPayTradeQueryResponse) PaymentResult {
	return PaymentResult{
		PaymentState:     r.AliPayTradeQuery.TradeStatus,
		PaymentStateDesc: "",
		Amount:           mustPriceToCent(r.AliPayTradeQuery.TotalAmount),
		TransactionID:    r.AliPayTradeQuery.TradeNo,
		OrderID:          r.AliPayTradeQuery.OutTradeNo,
		ConfirmedAt:      chrono.Time{},
	}
}

// NewPaymentResultAli builds PaymentResult from alipay webhook notification.
func NewPaymentResultAli(n *alipay.TradeNotification) (PaymentResult, error) {
	a, err := priceToCent(n.TotalAmount)
	if err != nil {
		return PaymentResult{}, err
	}

	return PaymentResult{
		Amount:      a,
		OrderID:     n.OutTradeNo,
		ConfirmedAt: chrono.TimeFrom(dt.MustParseAliTime(n.GmtPayment)),
	}, nil
}

// NewPaymentResultWx builds PaymentResult from wechat pay webhook notification.
func NewPaymentResultWx(n wechat.Notification) (PaymentResult, error) {
	if n.TotalFee.IsZero() {
		return PaymentResult{}, errors.New("no payment amount found in wx webhook")
	}

	if n.FTCOrderID.IsZero() {
		return PaymentResult{}, errors.New("no order id in wx webhook")
	}

	return PaymentResult{
		Amount:      n.TotalFee.Int64,
		OrderID:     n.FTCOrderID.String,
		ConfirmedAt: chrono.TimeFrom(dt.MustParseWxTime(n.TimeEnd.String)),
	}, nil
}
