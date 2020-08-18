package subscription

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/models/plan"
	"github.com/FTChinese/subscription-api/pkg/subs"
)

// PaymentIntent contains the data describing user's intent to pay.
// The data are constructed prior to payment.
type PaymentIntent struct {
	subs.Charge   // How much user should pay.
	subs.Duration // How long the membership this payment purchased.

	SubsKind enum.OrderKind `json:"subscriptionKind"`
	Wallet   subs.Wallet    `json:"wallet"`
	Plan     plan.Plan      `json:"plan"` // The plan to subscribe.
}
