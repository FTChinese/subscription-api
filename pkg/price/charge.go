package price

import (
	"github.com/FTChinese/subscription-api/pkg/conv"
)

// Charge represents how much we should ask user to pay in what currency with which discount.
type Charge struct {
	// The actual amount payable.
	Amount   float64 `json:"amount" db:"charged_amount"` // Actual price paid.
	Currency string  `json:"currency" db:"currency"`     // in which currency.
}

// NewCharge calculates the final price user should pay.
func NewCharge(price Price, offer Discount) Charge {
	// This is a temporary solution to avoid float precision problem.
	return Charge{
		Amount:   (price.UnitAmount*100 - offer.PriceOff.Float64*100) / 100,
		Currency: string(price.Currency),
	}
}

// AliPrice converts Charged price to ailpay format
// Deprecated.
func (c Charge) AliPrice() string {
	return conv.FormatMoney(c.Amount)
}

// AmountInCent converts Charged price to int64 in cent for comparison with wx notification.
// Deprecated
func (c Charge) AmountInCent() int64 {
	return int64(c.Amount * 100)
}
