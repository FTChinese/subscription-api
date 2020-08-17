package subscription

import (
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/models/plan"
)

// PaymentIntent contains the data describing user's intent to pay.
// The data are constructed prior to payment.
type PaymentIntent struct {
	Charge   // How much user should pay.
	Duration // How long the membership this payment purchased.

	SubsKind enum.OrderKind `json:"subscriptionKind"`
	Wallet   Wallet         `json:"wallet"`
	Plan     plan.Plan      `json:"plan"` // The plan to subscribe.
}
