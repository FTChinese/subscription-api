package product

import (
	"fmt"
	"strconv"
	"strings"
)

type Charge struct {
	// The actual amount payable.
	Amount   float64 `json:"amount" db:"charged_amount"` // Actual price paid.
	Currency string  `json:"currency" db:"currency"`     // in which currency.
}

// AliPrice converts Charged price to ailpay format
func (c Charge) AliPrice(live bool) string {
	if live {
		return strconv.FormatFloat(c.Amount, 'f', 2, 32)
	}

	return strconv.FormatFloat(0.01, 'f', 2, 32)
}

// AmountInCent converts Charged price to int64 in cent for comparison with wx notification.
func (c Charge) AmountInCent(live bool) int64 {
	if live {
		return int64(c.Amount * 100)
	}

	return 1
}

// ReadableAmount turns the charged amount to human readable format.
func (c Charge) ReadableAmount() string {
	return fmt.Sprintf("%s%.2f",
		strings.ToUpper(c.Currency),
		c.Amount,
	)
}
