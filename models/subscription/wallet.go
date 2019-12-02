package subscription

import (
	"github.com/FTChinese/go-rest/chrono"
	"math"
	"time"
)

// Wallet show how much money a member still owns and
// which orders constitutes the balance.
type Wallet struct {
	Balance       float64         `json:"balance"`
	Source        []ProratedOrder `json:"-"`
	EffectiveDate chrono.Time     `json:"effective_date"` // When the balance is calculated.
}

func NewWallet(orders []ProratedOrder, asOf time.Time) Wallet {
	// Time precision to date.
	asOf = asOf.Truncate(24 * time.Hour)

	w := Wallet{
		Balance:       0,
		Source:        []ProratedOrder{},
		EffectiveDate: chrono.TimeFrom(asOf),
	}

	for _, v := range orders {
		v.Balance = v.Prorate(asOf)
		w.Source = append(w.Source, v)

		// Aggregate each order's balance.
		w.Balance = w.Balance + v.Balance
	}

	return w
}

// ProratedOrder is used to retrieve paid order with balance.
type ProratedOrder struct {
	OrderID   string      `db:"order_id"`
	Amount    float64     `db:"amount"`
	StartDate chrono.Date `db:"start_date"`
	EndDate   chrono.Date `db:"end_date"`
	Balance   float64     `db:"balance"` // For insert only.
}

// Balance calculates the unused portion of an order up to
// current moment.
func (p ProratedOrder) Prorate(asOf time.Time) float64 {

	// If subscription starting date of this order is in
	// future, returns all the paid amount.
	// If the plan's start date is not passed yet, do not
	// perform calculation.
	if asOf.Before(p.StartDate.Time) {
		return p.Amount
	}

	// If start date is before today, it means portion of
	// this order has already been used.
	// Calculate the remaining portion.
	left := p.EndDate.Sub(asOf)

	// The duration this order purchased.
	total := p.EndDate.Sub(p.StartDate.Time)

	remains := left.Hours() * p.Amount / total.Hours()

	// If remains < 1, the result will be 1.0
	if remains < 1 {
		return remains
	}

	return math.Ceil(remains)
}
