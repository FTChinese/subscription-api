package paywall

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"gitlab.com/ftchinese/subscription-api/models/rand"
	"log"
	"math"
	"time"
)

// ProrationSource gets the unused portion of an order.
type ProrationSource struct {
	OrderID    string      `db:"order_id"`
	PaidAmount float64     `db:"paid_amount"`
	StartDate  chrono.Date `db:"start_date"`
	EndDate    chrono.Date `db:"end_date"`
	// The above fields are retrieved from db.
	Balance     float64     `db:"balance"`
	CreatedUTC  chrono.Time `db:"created_at"`  // The moment this record is created.
	ConsumedUTC chrono.Time `db:"consumed_at"` // The moment the upgrading order is confirmed.
	UpgradeID   string      `db:"upgrade_id"`
}

// Balance calculates the unused portion of an order up to
// current moment.
func (p ProrationSource) Prorate() float64 {
	today := time.Now().Truncate(24 * time.Hour)

	// If subscription starting date of this order is in
	// future, returns all the paid amount.
	// If the plan's start date is not passed yet, do not
	// perform calculation.
	if today.Before(p.StartDate.Time) {
		return p.PaidAmount
	}

	// If start date is before today, it means portion of
	// this order has already been used.
	// Calculate the remaining portion.
	left := p.EndDate.Sub(today)

	log.Printf("Left %+v", left)

	total := p.EndDate.Sub(p.StartDate.Time)

	log.Printf("Total %+v", total)

	remains := left.Hours() * p.PaidAmount / total.Hours()

	log.Printf("Remains %+v", remains)

	// If remains < 1, the result will be 1.0
	if remains < 1 {
		return remains
	}

	return math.Ceil(remains)
}

func GenerateUpgradeID() string {
	return "up_" + rand.String(12)
}

type UpgradePlan struct {
	ID string `json:"id" db:"upgrade_id"`
	//OrderID     null.String `json:"-" db:"order_id"`
	Balance float64 `json:"balance" db:"balance"` // Accumulated on all BalanceSource.Balance
	//SourceIDs   []string    `json:"sources" db:"source_id"` // The order ids which still have portion of days unused.
	CreatedAt chrono.Time `json:"createdAt" db:"created_at"`
	//ConfirmedAt chrono.Time `json:"-" db:"confirmed_at"`
	Data []ProrationSource `json:"data"`
	Plan Plan              `json:"plan"`
}

func NewUpgradePlan(sources []ProrationSource) UpgradePlan {
	up := UpgradePlan{
		ID: GenerateUpgradeID(),
		//SourceIDs: []string{},
		CreatedAt: chrono.TimeNow(),
		Data:      []ProrationSource{}, // This initializes to an empty array, rather than nil so that when marshaling into JSON, it is [], not null.
	}

	for _, v := range sources {
		v.Balance = v.Prorate()
		up.Balance = up.Balance + v.Balance
		//up.SourceIDs = append(up.SourceIDs, v.OrderID)
		v.UpgradeID = up.ID
		up.Data = append(up.Data, v)
	}

	// This is hardcoded. Should refactor in the future.
	up.Plan = premiumYearlyPlan.WithUpgrade(up.Balance)
	return up
}

func (up UpgradePlan) ReadableBalance() string {
	return fmt.Sprintf("CNY%.2f", up.Balance)
}
