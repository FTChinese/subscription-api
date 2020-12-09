package stripe

import (
	"github.com/stripe/stripe-go/v72"
)

var subStatusLocalized = map[stripe.SubscriptionStatus]string{
	stripe.SubscriptionStatusActive:            "活跃",
	stripe.SubscriptionStatusIncomplete:        "支付未完成，请在24小时内完成支付",
	stripe.SubscriptionStatusIncompleteExpired: "支付已过期",
	stripe.SubscriptionStatusPastDue:           "续费失败",
	stripe.SubscriptionStatusCanceled:          "订阅已取消",
	stripe.SubscriptionStatusUnpaid:            "Stripe未能找到合适的支付方式，支付已取消",
}

// LocalizeStripeSubStatus translate stripe subscription status to Chinese.
func LocalizeStripeSubStatus(st stripe.SubscriptionStatus) string {
	s, ok := subStatusLocalized[st]
	if !ok {
		return string(st)
	}

	return s
}

//See https://stripe.com/docs/billing/invoices/workflow#workflow-overview
var invoiceStatusLocalized = map[stripe.InvoiceStatus]string{
	stripe.InvoiceStatusDraft:         "草稿",
	stripe.InvoiceStatusOpen:          "等待支付",
	stripe.InvoiceStatusPaid:          "已支付",
	stripe.InvoiceStatusUncollectible: "无法收款",
	stripe.InvoiceStatusVoid:          "错误，应撤销",
}

func LocalizeStripeInvoiceStatus(st stripe.InvoiceStatus) string {
	s, ok := invoiceStatusLocalized[st]
	if !ok {
		return string(st)
	}

	return s
}
