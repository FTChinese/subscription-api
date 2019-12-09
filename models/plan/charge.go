package plan

import "strconv"

type Charge struct {
	Amount   float64 `json:"amount" db:"amount"`     // Actual price paid.
	Currency string  `json:"currency" db:"currency"` // in which currency.
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
