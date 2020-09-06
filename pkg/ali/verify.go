package ali

import (
	"github.com/smartwalle/alipay"
)

type TradeStatus string

const (
	TradeStatusFinished TradeStatus = "TRADE_FINISHED"
	TradeStatusSuccess  TradeStatus = "TRADE_SUCCESS"
	TradeStatusPending  TradeStatus = "WAIT_BUYER_PAY"
	TradeStatusClosed   TradeStatus = "TRADE_CLOSED"

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
