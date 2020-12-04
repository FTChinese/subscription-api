package stripe

import (
	"github.com/stripe/stripe-go/v71"
)

var statusLocalized = map[stripe.SubscriptionStatus]string{
	stripe.SubscriptionStatusActive:            "活跃",
	stripe.SubscriptionStatusIncomplete:        "支付未完成，请在24小时内完成支付",
	stripe.SubscriptionStatusIncompleteExpired: "支付已过期",
	stripe.SubscriptionStatusPastDue:           "续费失败",
	stripe.SubscriptionStatusCanceled:          "订阅已取消",
	stripe.SubscriptionStatusUnpaid:            "Stripe未能找到合适的支付方式，支付已取消",
}

// LocalizeStripeStatus translate stripe subscription status to Chinese.
func LocalizeStripeStatus(st stripe.SubscriptionStatus) string {
	s, ok := statusLocalized[st]
	if !ok {
		return string(st)
	}

	return s
}
