package subscription

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/rand"
)

func GenerateUpgradeID() string {
	return "up_" + rand.String(12)
}

// UpgradeBalanceSchema records the balance the moment user
// starts upgrading and the upgraded plan.
type UpgradeBalanceSchema struct {
	ID         string      `db:"upgrade_id"`
	Balance    float64     `db:"balance"`    // Wallet.Balance
	CreatedAt  chrono.Time `db:"created_at"` // Wallet.AsOfDate
	PlanPrice  float64     `db:"plan_price"`
	PlanAmount float64     `db:"plan_amount"`
}

// BuildWallet turns the upgrading balance data back to wallet.
func (u UpgradeBalanceSchema) BuildWallet(orders []ProratedOrderSchema) Wallet {
	sources := make([]ProratedOrder, 0)

	for _, v := range orders {
		sources = append(sources, v.ProratedOrder)
	}

	return Wallet{
		Balance:  u.Balance,
		Source:   sources,
		AsOfDate: u.CreatedAt,
	}
}

// ReadableBalance produces a string describing the total balance
// in the format: CNY99.00 in email sent to user.
// Deprecate
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
