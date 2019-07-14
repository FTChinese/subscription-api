package paywall

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"math"
	"strings"
	"time"
)

func GenerateUpgradeID() (string, error) {
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
	Payable     float64     `json:"payable"` // Deprecate. The amount user needs to pay.
	Source      []string    `json:"-"`       // The order ids which still have portion of days unused.
	CreatedAt   chrono.Time `json:"-"`
	ConfirmedAt chrono.Time `json:"-"`
	Member      Membership  `json:"-"` // Membership status prior to upgrade.
}

// NewUpgrade creates an Upgrade instance based on the plan selected.
// Actually the only plan available is premium_year.
func NewUpgrade(p Plan) Upgrade {
	id, _ := GenerateUpgradeID()
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
func (p Upgrade) SetBalance(sources []BalanceSource) Upgrade {
	for _, v := range sources {
		p.Balance = p.Balance + v.Balance()
		p.Source = append(p.Source, v.ID)
	}

	return p
}

// CalculatePrice determines how user should pay for an upgrade.
func (p Upgrade) CalculatePayable() Upgrade {
	// Is Balance big enough to cover NetPrice.

	if p.ListPrice > p.Balance {
		// UserID should pay diff
		p.NetPrice = p.ListPrice - p.Balance
		p.Payable = p.NetPrice
		p.CycleCount = 1
		p.ExtraDays = 1
	} else {
		// Enough to cover the gap. User do not need to pay.
		p.CycleCount, p.ExtraDays = convertBalance(p.Balance, p.Plan.ListPrice)

		p.Payable = 0
		p.NetPrice = 0
	}

	return p
}

func convertBalance(balance, price float64) (int64, int64) {
	var cycles int64 = 0

	for balance > price {
		cycles = cycles + 1
		balance = balance - price
	}

	days := math.Ceil(balance * 365 / price)

	return cycles, int64(days)
}

type BalanceSource struct {
	ID        string // Order id.
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
