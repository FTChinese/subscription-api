package ali

import (
	"github.com/smartwalle/alipay"
)

const (
	SubCodeSysErr       = "ACQ.SYSTEM_ERROR"
	SubCodeInvalidParam = "ACQ.INVALID_PARAMETER"
	SubCodeNotExists    = "ACQ.TRADE_NOT_EXIST"
)

const (
	TradeStatusFinished = "TRADE_FINISHED"
	TradeStatusSuccess  = "TRADE_SUCCESS"
	TradeStatusPending  = "WAIT_BUYER_PAY"
	TradeStatusClosed   = "TRADE_CLOSED"
)

var TradeStatusMsg = map[string]string{
	TradeStatusPending:  "交易创建，等待付款",
	TradeStatusClosed:   "交易关闭",
	TradeStatusSuccess:  "交易支付成功",
	TradeStatusFinished: "交易结束",
}

func IsPaySuccess(n *alipay.TradeNotification) bool {
	switch n.TradeStatus {
	case TradeStatusSuccess, TradeStatusFinished:
		return true
	default:
		return false
	}
}

func IsStatusSuccess(s string) bool {
	switch s {
	case TradeStatusSuccess, TradeStatusFinished:
		return true
	default:
		return false
	}
}

func ShouldRetry(n *alipay.TradeNotification) bool {
	switch n.TradeStatus {
	case TradeStatusPending:
		return true
	default:
		return false
	}
}
