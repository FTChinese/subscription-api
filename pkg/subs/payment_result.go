package subs

import (
	"errors"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/smartwalle/alipay"
	"strconv"
	"time"
)

// PaymentResult unifies ali and wx webhook notification.
type PaymentResult struct {
	Amount      int64 // Unify various payment amounts to cents.
	OrderID     string
	ConfirmedAt time.Time
}

func NewPaymentResultAli(n *alipay.TradeNotification) (PaymentResult, error) {
	f, err := strconv.ParseFloat(n.TotalAmount, 64)
	if err != nil {
		return PaymentResult{}, err
	}

	return PaymentResult{
		Amount:      int64(f * 100),
		OrderID:     n.OutTradeNo,
		ConfirmedAt: ParseAliTime(n.GmtPayment),
	}, nil
}

func NewPaymentResultWx(n wechat.Notification) (PaymentResult, error) {
	if n.TotalFee.IsZero() {
		return PaymentResult{}, errors.New("no payment amount found in wx webhook")
	}

	if n.FTCOrderID.IsZero() {
		return PaymentResult{}, errors.New("no order id in wx webhook")
	}

	confirmedAt, err := ParseWxTime(n.TimeEnd.String)
	if err != nil {
		confirmedAt = time.Now()
	}

	return PaymentResult{
		Amount:      n.TotalFee.Int64,
		OrderID:     n.FTCOrderID.String,
		ConfirmedAt: confirmedAt,
	}, nil
}
