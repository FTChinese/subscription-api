package wechat

import (
	"github.com/objcoding/wxpay"
)

// WebhookResult contains the essential data of a webhook payload.
type WebhookResult struct {
	ResultCode    string // 业务结果
	ErrorMessage  string // 错误代码描述
	FTCOrderID    string
	TotalFee      int64
	TransactionID string
	TimeEnd       string
}

// NewWebhookParams converts wxpay.Params type to WebhookResult type.
func NewWebhookParams(p wxpay.Params) WebhookResult {
	return WebhookResult{
		ResultCode:    p.GetString(keyResultCode),
		ErrorMessage:  p.GetString(keyErrCodeDes),
		FTCOrderID:    p.GetString(keyOrderID),
		TotalFee:      p.GetInt64(keyTotalAmount),
		TransactionID: p.GetString(keyTxnID),
		TimeEnd:       p.GetString(keyEndTime),
	}
}
