package subs

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
)

// UpgradeIntent contains the data describing user's intent to pay.
// The data are constructed prior to payment.
type UpgradeIntent struct {
	product.Charge                      // Deprecated. How much user should pay.
	CycleCount     int64                `json:"cycleCount"`       // Deprecated
	ExtraDays      int64                `json:"extraDays"`        // Deprecated
	LegacySubsKind enum.OrderKind       `json:"subscriptionKind"` // Deprecated
	Plan           product.ExpandedPlan `json:"plan"`             // Deprecated. The plan to subscribe.

	Checkout
	Result ConfirmationResult `json:"-"` // Only for free upgrade.
}
