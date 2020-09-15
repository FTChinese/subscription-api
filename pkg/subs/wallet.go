package subs

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/arith"
	"github.com/FTChinese/subscription-api/pkg/product"
	"math"
	"time"
)

// ProratedOrder is the balance of an order upon upgrading.
type ProratedOrder struct {
	OrderID        string      `db:"order_id"` // Balance source order's order.
	Balance        float64     `db:"balance"`
	CreatedUTC     chrono.Time `db:"created_utc"`
	ConsumedUTC    chrono.Time `db:"consumed_utc"`
	UpgradeOrderID string      `db:"upgrade_order_id"` // The order id to perform upgrade.
}

// BalanceSource is used to retrieve confirmed orders with unused portion.
type BalanceSource struct {
	OrderID   string      `db:"order_id"`
	Amount    float64     `db:"charged_amount"` // Retrieve only
	StartDate chrono.Date `db:"start_date"`     // Retrieval only
	EndDate   chrono.Date `db:"end_date"`       // Retrieval only
}

// GetBalance calculates the unused portion of an order up to
// current moment.
func (b BalanceSource) GetBalance(asOf time.Time) float64 {

	// If subscription starting date of this order is in
	// future, returns all the paid amount.
	// If the plan's start date is not passed yet, do not
	// perform calculation.
	if asOf.Before(b.StartDate.Time) {
		return b.Amount
	}

	// If start date is before today, it means portion of
	// this order has already been used.
	// Calculate the remaining portion.
	left := b.EndDate.Sub(asOf)

	// The duration this order purchased.
	total := b.EndDate.Sub(b.StartDate.Time)

	remains := left.Hours() * b.Amount / total.Hours()

	// If remains < 1, the result will be 1.0
	if remains < 1 {
		return remains
	}

	return math.Ceil(remains)
}

// Prorated builds a ProratedOrder.
// `asOf` is the ending time to calculate balance.
func (b BalanceSource) Prorated(asOf time.Time) ProratedOrder {
	return ProratedOrder{
		OrderID:        b.OrderID,
		Balance:        b.GetBalance(asOf),
		CreatedUTC:     chrono.TimeNow(),
		ConsumedUTC:    chrono.Time{},
		UpgradeOrderID: "",
	}
}

// Wallet show how much money a member still owns and
// which orders constitutes the balance.
// Wallet is dynamic and changes as time passed.
type Wallet struct {
	Balance   float64         `json:"balance" db:"balance"`
	CreatedAt chrono.Time     `json:"createdAt" db:"created_utc"`
	Sources   []ProratedOrder `json:"-"`
}

// NewWallet creates a wallet based on a user's orders which have portion not used yet.
// The passed in orders will have its balance calculated as of now.
// A wallet's balance is the sum of each order's remaining balance.
func NewWallet(sources []BalanceSource, asOf time.Time) Wallet {
	// Time precision to date.
	asOf = asOf.Truncate(24 * time.Hour)

	w := Wallet{
		Balance:   0,
		CreatedAt: chrono.TimeFrom(asOf),
		Sources:   make([]ProratedOrder, 0),
	}

	// Calculate each order's balance.
	for _, v := range sources {
		po := v.Prorated(asOf)

		w.Sources = append(w.Sources, v.Prorated(asOf))

		// Aggregate each order's balance.
		w.Balance = w.Balance + po.Balance
	}

	return w
}

func (w Wallet) ReadableBalance() string {
	return fmt.Sprintf("%s%.2f", "CNY", w.Balance)
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

func (w Wallet) CheckOut(p product.ExpandedPlan) Checkout {
	dur := w.ConvertBalance(p)

	charge := p.Payable()

	// If user's wallet has balance.
	if w.Balance > 0 {
		charge.Amount = charge.Amount - w.Balance
	}

	// If balance exceeds payable amount
	if charge.Amount < 0 {
		charge.Amount = 0
	}

	return Checkout{
		Charge:   charge,
		Duration: dur,
	}
}

// WithUpgradeOrder updates the wallet's sources after upgrading order created.
func (w Wallet) WithUpgradeOrder(o Order) Wallet {
	isFreeUpgrade := o.IsFreeUpgrade()
	now := chrono.TimeNow()

	for i, v := range w.Sources {
		v.UpgradeOrderID = o.ID
		if isFreeUpgrade {
			v.ConsumedUTC = now
		}

		w.Sources[i] = v
	}

	return w
}
