package subs

import (
	"errors"
	"github.com/FTChinese/subscription-api/pkg/dt"
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

// NewPaymentResultAli builds PaymentResult from alipay webhook notification.
func NewPaymentResultAli(n *alipay.TradeNotification) (PaymentResult, error) {
	f, err := strconv.ParseFloat(n.TotalAmount, 64)
	if err != nil {
		return PaymentResult{}, err
	}

	return PaymentResult{
		Amount:      int64(f * 100),
		OrderID:     n.OutTradeNo,
		ConfirmedAt: dt.MustParseAliTime(n.GmtPayment),
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
		ConfirmedAt: dt.MustParseWxTime(n.TimeEnd.String),
	}, nil
}
