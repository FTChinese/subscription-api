package subs

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
)

// PaymentIntent contains the data describing user's intent to pay.
// The data are constructed prior to payment.
type PaymentIntent struct {
	product.Charge   // How much user should pay.
	product.Duration // How long the membership this payment purchased.

	SubsKind enum.OrderKind     `json:"subscriptionKind"`
	Wallet   Wallet             `json:"wallet"`
	Plan     product.IntentPlan `json:"plan"` // The plan to subscribe.
}
