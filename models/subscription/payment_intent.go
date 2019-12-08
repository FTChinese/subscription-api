package subscription

import (
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"strconv"
)

type Charge struct {
	Amount     float64 // Actual price paid.
	Currency   string  // in which currency.
}

// AliPrice converts Charged price to ailpay format
func (c Charge) AliPrice(sandbox bool) string {
	if sandbox {
		return strconv.FormatFloat(0.01, 'f', 2, 32)
	}

	return strconv.FormatFloat(c.Amount, 'f', 2, 32)
}

// AmountInCent converts Charged price to int64 in cent for comparison with wx notification.
func (c Charge) AmountInCent(sandbox bool) int64 {
	if sandbox {
		return 1
	}
	return int64(c.Amount * 100)
}

type PaymentIntent struct {
	Charge
	plan.Duration

	SubsKind plan.SubsKind `json:"subscriptionKind"`
	Wallet   Wallet        `json:"wallet"`
	Plan     plan.Plan     `json:"plan"`
}
