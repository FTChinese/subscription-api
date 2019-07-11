package paywall

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"math"
	"strings"
	"time"
)

func genUpgradeID() (string, error) {
	s, err := gorest.RandomBase64(9)
	if err != nil {
		return "", err
	}

	return "up_" + s, nil
}

type Upgrade struct {
	ID string `json:"id"`
	Plan
	Balance     float64     `json:"balance"` // Accumulated on all BalanceSource.Balance
	Payable     float64     `json:"payable"` // The amount user needs to pay.
	Source      []string    `json:"-"`       // The order ids which still have portion of days unused.
	CreatedAt   chrono.Time `json:"createdAt"`
	ConfirmedAt chrono.Time `json:"confirmedAt"`
	Member      Membership  `json:"-"` // Membership status prior to upgrade.
}

// NewUpgrade creates an Upgrade instance based on the plan selected.
// Actually the only plan available is premium_year.
func NewUpgrade(p Plan) Upgrade {
	id, _ := genUpgradeID()
	return Upgrade{
		ID:   id,
		Plan: p,
	}
}

// SourceOrderIDs concatenate balance source order id into a
// string so that it could be inserted into db.
func (p Upgrade) SourceOrderIDs() string {
	return strings.Join(p.Source, ",")
}

// SetBalance sets the balance for an upgrade and where those
// balances comes from.
func (p Upgrade) SetBalance(orders []BalanceSource) Upgrade {
	for _, v := range orders {
		p.Balance = p.Balance + v.Balance()
		p.Source = append(p.Source, v.ID)
	}

	return p
}

// CalculatePrice determines how user should pay for an upgrade.
func (p Upgrade) CalculatePayable() Upgrade {
	// Is Balance big enough to cover NetPrice.
	diff := p.ListPrice - p.Balance

	if diff >= 0 {
		// User should pay diff
		p.NetPrice = diff
		p.Payable = diff
		p.CycleCount = 1
		p.ExtraDays = 1
	} else {
		// Enough to cover the gap. User do not need to pay.
		p.Payable = 0
		p.NetPrice = 0

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

type BalanceSource struct {
	ID        string
	NetPrice  float64
	StartDate chrono.Date
	EndDate   chrono.Date
}

// Balance calculates the unused portion of an order
func (s BalanceSource) Balance() float64 {
	today := time.Now().Truncate(24 * time.Hour)

	// If subscription starting date of this order is in
	// future, returns all the paid amount.
	if !s.StartDate.Before(today) {
		return s.NetPrice
	}

	// If start date is before today, it means portion of
	// this order has already been used.
	// Calculate the remaining portion.
	left := s.EndDate.Sub(today)

	total := s.EndDate.Sub(s.StartDate.Time)

	remains := left.Hours() * s.NetPrice / total.Hours()

	return math.Ceil(remains)
}
