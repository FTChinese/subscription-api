package subscription

import (
	"gitlab.com/ftchinese/subscription-api/models/plan"
)

type PaymentIntent struct {
	plan.Charge
	Duration

	SubsKind plan.SubsKind `json:"subscriptionKind"`
	Wallet   Wallet        `json:"wallet"`
	Plan     plan.Plan     `json:"plan"`
}
