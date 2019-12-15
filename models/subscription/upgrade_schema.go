package subscription

import (
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
// This is actually a snapshot of user's wallet persisted in DB, plus the plan upgraded to.
type UpgradeSchema struct {
	ID string `db:"upgrade_id"`
	BaseWallet
	plan.Plan
	Sources []ProratedOrderSchema // The existing orders not consumed completely yet.
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
