package paywall

import (
	"github.com/FTChinese/go-rest/chrono"
	"math"
	"time"
)

type UpgradePlan struct {
	Plan
	Balance    float64  `json:"balance"` // Accumulated on all Proration.Balance
	CycleCount int64    `json:"cycleCount"`
	ExtraDays  int64    `json:"extraDays"` // If Balance is larger than the price to cover one cycle.
	Payable    float64  `json:"payable"`   // The amount user needs to pay.
	OrderIDs   []string `json:"-"`         // From which orders you get Balance
}

func NewUpgradePlan(p Plan) UpgradePlan {
	up := UpgradePlan{}
	up.Tier = p.Tier
	up.Cycle = p.Cycle
	up.ListPrice = p.ListPrice
	up.NetPrice = p.NetPrice
	up.Description = p.Description

	return up
}

func (p UpgradePlan) SetBalance(orders []UnusedOrder) UpgradePlan {
	for _, v := range orders {
		p.Balance = p.Balance + v.Balance()
		p.OrderIDs = append(p.OrderIDs, v.ID)
	}

	return p
}

// CalculatePrice determines how user should pay for an upgrade.
func (p UpgradePlan) CalculatePayable() UpgradePlan {
	// Is Balance big enough to cover NetPrice.
	diff := p.NetPrice - p.Balance

	if diff >= 0 {
		// User should pay diff
		p.Payable = diff
		p.CycleCount = 1
		p.ExtraDays = 1
	} else {
		// Enough to cover the gap. User do not need to pay.
		p.Payable = 0

		// The balance might be multiple of `price`
		quotient := p.Balance / p.NetPrice
		// Balance is larger than `price`
		// Example 3000.0 / 1998.0 = 1.5015015015015014
		p.CycleCount = int64(math.Trunc(quotient))

		// Change remaining amount to days
		// Example: math.Mod(3000.0, 1998.0) = 1002
		// It means there are $1002 left in user pocket.
		// But $1002 cannot be changed for a 1 year's membership
		r := math.Mod(p.Balance, p.ListPrice)

		// Then see how many days could the remainder be changed.
		// Example: 1002 * 365 / 1998.0 = 183.04804804804806
		// Use math.Ceil to give user a bonus day.
		days := math.Ceil(r * 365 / p.NetPrice)

		p.ExtraDays = int64(days)
	}

	return p
}

// Proration is an order that is in use or not used yet.
//type Proration struct {
//	OrderID   string
//	Balance   float64
//	StartDate chrono.Date
//	EndDate   chrono.Date
//}

type UnusedOrder struct {
	ID        string
	NetPrice  float64
	StartDate chrono.Date
	EndDate   chrono.Date
}

// Balance calculates the unused portion of an order
func (o UnusedOrder) Balance() float64 {
	today := time.Now().Truncate(24 * time.Hour)

	// If subscription starting date of this order is in
	// future, returns all the paid amount.
	if !o.StartDate.Before(today) {
		return o.NetPrice
	}

	// If start date is before today, it means portion of
	// this order has already been used.
	// Calculate the remaining portion.
	left := o.EndDate.Sub(today)

	total := o.EndDate.Sub(o.StartDate.Time)

	remains := left.Hours() * o.NetPrice / total.Hours()

	return math.Ceil(remains)
}
