package price

import (
	"strconv"
)

// Charge represents how much we should ask user to pay in what currency with which discount.
type Charge struct {
	// The actual amount payable.
	Amount   float64 `json:"amount" db:"charged_amount"` // Actual price paid.
	Currency string  `json:"currency" db:"currency"`     // in which currency.
}

func NewCharge(price Price, offer Discount) Charge {
	return Charge{
		Amount:   price.UnitAmount - offer.PriceOff.Float64,
		Currency: string(price.Currency),
	}
}

// AliPrice converts Charged price to ailpay format
func (c Charge) AliPrice() string {
	return strconv.FormatFloat(c.Amount, 'f', 2, 32)
}

// AmountInCent converts Charged price to int64 in cent for comparison with wx notification.
func (c Charge) AmountInCent() int64 {
	return int64(c.Amount * 100)
}

func FormatMoney(n float64) string {
	return strconv.FormatFloat(n, 'f', 2, 32)
}
