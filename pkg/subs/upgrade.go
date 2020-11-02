package subs

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
)

// UpgradeIntent contains the data describing user's intent to pay.
// The data are constructed prior to payment.
type UpgradeIntent struct {
	product.Charge                  // Deprecated. How much user should pay.
	product.Duration                // Deprecated. How long the membership this payment purchased.
	LegacySubsKind   enum.OrderKind `json:"subscriptionKind"` // Deprecated

	SubsKind enum.OrderKind       `json:"kind"`
	Plan     product.ExpandedPlan `json:"plan"` // The plan to subscribe.
	Discount product.Discount     `json:"discount"`
	Wallet   Wallet               `json:"wallet"`
	Payable  product.Charge       `json:"payable"`
	Length   product.Duration     `json:"duration"`
	IsFree   bool                 `json:"isFree"`
	Result   ConfirmationResult   `json:"-"` // Only for free upgrade.
}
