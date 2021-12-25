package price

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
