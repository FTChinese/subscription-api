package plan

type PaymentIntent struct {
	Price      float64 `json:"price" db:"price"`
	Amount     float64 `json:"amount" db:"amout"`
	Currency   string  `json:"currency" db:"currency"`
	Title      string  `json:"-"`
	CycleCount int64   `json:"cycle_count" db:"cycle_count"`
	ExtraDays  int64   `json:"extra_days" db:"extra_days"`
}
