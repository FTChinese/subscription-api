package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/guregu/null"
)

// UpgradeSchema mimics SQL structure of storing
// upgrading info.
// The are split into two tables:
// one row of UpgradeBalanceSchema maps to multiple
// rows ProratedOrderSchema.
// This is actually a snapshot of user's wallet persisted in DB, plus the plan upgraded to.
type UpgradeSchema struct {
	ID           string                `db:"upgrade_id"`
	Balance      float64               `db:"balance"`
	CreatedUtc   chrono.Time           `db:"created_utc"`
	OrderID      null.String           `db:"order_id"`
	Sources      []ProratedOrderSchema // The existing orders not consumed completely yet.
	product.Plan                       // Deprecated as it is already saved in order table.
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
	CreatedUTC  chrono.Time `db:"created_utc"`  // The moment this record is created. Retrieval only
	ConsumedUTC chrono.Time `db:"consumed_utc"` // The moment the upgrading order is confirmed. Retrieval only.
	UpgradeID   string      `db:"upgrade_id"`
}
