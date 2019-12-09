package subscription

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/rand"
)

func GenerateUpgradeID() string {
	return "up_" + rand.String(12)
}

// UpgradeSchema specifies how a standard product is upgraded to premium.
type UpgradeSchema struct {
	ID         string      `db:"upgrade_id"`
	Balance    float64     `db:"balance"`
	CreatedAt  chrono.Time `db:"created_at"`
	PlanPrice  float64     `db:"plan_price"`
	PlanAmount float64     `db:"plan_amount"`
}

// ReadableBalance produces a string describing the total balance
// in the format: CNY99.00 in email sent to user.
// Deprecate
func (up UpgradeSchema) ReadableBalance() string {
	return fmt.Sprintf("%s%.2f", "CNY", up.Balance)
}
