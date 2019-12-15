package subscription

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"math"
	"time"
)

type BaseWallet struct {
	Balance   float64     `json:"balance" db:"balance"`
	CreatedAt chrono.Time `json:"as_of_date" db:"created_at"` // When the balance is calculated.
}

func (w BaseWallet) ReadableBalance() string {
	return fmt.Sprintf("%s%.2f", "CNY", w.Balance)
}

// Wallet show how much money a member still owns and
// which orders constitutes the balance.
// Wallet is dynamic and changes as time passed.
type Wallet struct {
	BaseWallet
	Sources []ProratedOrder `json:"-"`
}

func NewWallet(orders []ProratedOrder, asOf time.Time) Wallet {
	// Time precision to date.
	asOf = asOf.Truncate(24 * time.Hour)

	w := Wallet{
		BaseWallet: BaseWallet{
			Balance:   0,
			CreatedAt: chrono.TimeFrom(asOf),
		},
		Sources: []ProratedOrder{},
	}

	for _, v := range orders {
		v.Balance = v.Prorate(asOf)
		w.Sources = append(w.Sources, v)

		// Aggregate each order's balance.
		w.Balance = w.Balance + v.Balance
	}

	return w
}

func (w Wallet) ConvertBalance(p plan.Plan) Duration {
	if w.Balance == 0 || w.Balance <= p.Price {
		return Duration{
			CycleCount: 1,
			ExtraDays:  1,
		}
	}

	// Balance is greater than product price.
	// User do not need to pay.
	// Convert balance directly to subscription period.
	cycles, remains := util.Division(w.Balance, p.Price)

	days := math.Ceil(remains * 365 / p.Price)

	return Duration{
		CycleCount: cycles,
		ExtraDays:  int64(days),
	}
}

// ProratedOrder is used to retrieve paid order with balance.
type ProratedOrder struct {
	OrderID   string      `db:"order_id"`
	Amount    float64     `db:"charged_amount"` // Retrieve only
	StartDate chrono.Date `db:"start_date"`     // Retrieval only
	EndDate   chrono.Date `db:"end_date"`       // Retrieval only
	Balance   float64     `db:"balance"`        // For insert only.
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

func (p ProratedOrder) ReadableBalance() string {
	return fmt.Sprintf("%s%.2f", "CNY", p.Balance)
}
