package ali

import (
	"github.com/smartwalle/alipay"
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

func ShouldRetry(n *alipay.TradeNotification) bool {
	switch n.TradeStatus {
	case tradePending:
		return true
	default:
		return false
	}
}
