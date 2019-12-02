package subscription

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"math"
	"time"
)

// ProrationSource gets the unused portion of an order.
// This is the balance of each valid order the moment user
// is requesting upgrade.
// Once the webhook received notification, each record
// will have ConsumedUTC set, indicating this order won't be
// included the nex time user requesting upgrade, which might
// happen if user stopped premium subscription, re-subscribed
// to standard product, and then upgrade again.
type ProrationSource struct {
	OrderID    string      `db:"order_id"`
	PaidAmount float64     `db:"paid_amount"`
	StartDate  chrono.Date `db:"start_date"`
	EndDate    chrono.Date `db:"end_date"`
	// The above fields are retrieved from db.
	Balance     float64     `db:"balance"`
	CreatedUTC  chrono.Time `db:"created_at"`  // The moment this record is created. Retrieval only
	ConsumedUTC chrono.Time `db:"consumed_at"` // The moment the upgrading order is confirmed. Retrieval only.
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

	total := p.EndDate.Sub(p.StartDate.Time)

	remains := left.Hours() * p.PaidAmount / total.Hours()

	// If remains < 1, the result will be 1.0
	if remains < 1 {
		return remains
	}

	return math.Ceil(remains)
}

func (p ProrationSource) ReadableBalance() string {
	return fmt.Sprintf("%s%.2f", "CNY", p.Balance)
}
