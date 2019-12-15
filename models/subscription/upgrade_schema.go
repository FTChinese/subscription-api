package subscription

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/rand"
	"gitlab.com/ftchinese/subscription-api/models/plan"
)

func GenerateUpgradeID() string {
	return "up_" + rand.String(12)
}

// UpgradeSchema mimics SQL structure of storing
// upgrading info.
// The are split into two tables:
// one row of UpgradeBalanceSchema maps to multiple
// rows ProratedOrderSchema.
// This is actually a snapshot of user's wallet persisted in DB.
type UpgradeSchema struct {
	UpgradeBalanceSchema                       // Current balance
	Sources              []ProratedOrderSchema // The existing orders not consumed completely yet.
}

// UpgradeBalanceSchema records the balance the moment user
// starts upgrading and the upgraded plan.
// This is a snapshot of Wallet the moment user
// initialize upgrading.
type UpgradeBalanceSchema struct {
	ID        string      `db:"upgrade_id"`
	Balance   float64     `db:"balance"`    // Wallet.Balance
	CreatedAt chrono.Time `db:"created_at"` // Wallet.AsOfDate
	plan.Plan             // Save the plan's tier, cycle, price, amount and currency the moment user upgrades.
}

// ReadableBalance produces a string describing the total balance
// in the format: CNY99.00 in email sent to user.
func (u UpgradeBalanceSchema) ReadableBalance() string {
	return fmt.Sprintf("%s%.2f", "CNY", u.Balance)
}

// ProratedOrderSchema records the orders used to calculate balane.
// This is the balance of each valid order the moment user
// is requesting upgrade.
// Once the webhook received notification, each record
// will have ConsumedUTC set, indicating this order won't be
// included the next time user requesting upgrade, which might
// happen if user stopped premium subscription, re-subscribed
// to standard product, and then upgrade again.
type ProratedOrderSchema struct {
	ProratedOrder
	CreatedUTC  chrono.Time `db:"created_at"`  // The moment this record is created. Retrieval only
	ConsumedUTC chrono.Time `db:"consumed_at"` // The moment the upgrading order is confirmed. Retrieval only.
	UpgradeID   string      `db:"upgrade_id"`
}
