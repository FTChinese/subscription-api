package ali

import (
	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"strconv"
)

const (
	tradeFinished = "TRADE_FINISHED"
	tradeSuccess  = "TRADE_SUCCESS"
	tradePending  = "WAIT_BUYER_PAY"
	tradeClosed   = "TRADE_CLOSED"
)

func IsPaySuccess(n *alipay.TradeNotification) bool {
	switch n.TradeStatus {
	case tradeSuccess, tradeFinished:
		return true
	default:
		return false
	}
}

func GetPaymentResult(n *alipay.TradeNotification) (subscription.PaymentResult, error) {
	f, err := strconv.ParseFloat(n.TotalAmount, 64)
	if err != nil {
		return subscription.PaymentResult{}, err
	}

	return subscription.PaymentResult{
		Amount:      int64(f * 100),
		OrderID:     n.OutTradeNo,
		ConfirmedAt: util.ParseAliTime(n.GmtPayment),
	}, nil
}

func ShouldRetry(n *alipay.TradeNotification) bool {
	switch n.TradeStatus {
	case tradePending:
		return true
	default:
		return false
	}
}
