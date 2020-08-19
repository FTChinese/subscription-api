package subs

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/arith"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/guregu/null"
	"math"
	"time"
)

// ProratedOrder is used to retrieve confirmed orders with unused portion.
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

type BaseWallet struct {
	Balance   float64     `json:"balance" db:"balance"`
	CreatedAt chrono.Time `json:"createdAt" db:"created_utc"` // When the balance is calculated.
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

// NewWallet creates a wallet based on a user's orders which have portion not used yet.
// The passed in orders will have its balance calculated as of now.
// A wallet's balance is the sum of each order's remaining balance.
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

	// Calculate each order's balance.
	for _, v := range orders {
		v.Balance = v.Prorate(asOf)
		w.Sources = append(w.Sources, v)

		// Aggregate each order's balance.
		w.Balance = w.Balance + v.Balance
	}

	return w
}

// ConvertBalance calculates how many billing cycles could a user's current
// balance be converted to.
// If the balance is less thant a plan's price, returns the default duration;
// otherwise calculate
func (w Wallet) ConvertBalance(p product.ExpandedPlan) product.Duration {
	if w.Balance == 0 || w.Balance <= p.Price {
		return product.DefaultDuration()
	}

	// Balance is greater than product price.
	// User do not need to pay.
	// Convert balance directly to subscription period.
	cycles, remains := arith.Division(w.Balance, p.Price)

	days := math.Ceil(remains * 365 / p.Price)

	return product.Duration{
		CycleCount: cycles,
		ExtraDays:  int64(days),
	}
}

// UpgradeSchema converts wallet so that
// it could be used by sqlx to save data.
func (w Wallet) UpgradeSchema(orderID string, freeUpgrade bool) UpgradeSchema {
	id := GenerateUpgradeID()

	pos := make([]ProratedOrderSchema, 0)
	now := chrono.TimeNow()

	for _, v := range w.Sources {
		s := ProratedOrderSchema{
			ProratedOrder: v,
			CreatedUTC:    now,
			ConsumedUTC:   chrono.Time{},
			UpgradeID:     id,
		}

		if freeUpgrade {
			s.ConsumedUTC = now
		}

		pos = append(pos, s)
	}

	return UpgradeSchema{
		ID:         id,
		Balance:    w.Balance,
		CreatedUtc: w.CreatedAt,
		OrderID:    null.StringFrom(orderID),
		Sources:    pos,
	}
}
