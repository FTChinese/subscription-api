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

// UpgradeSchema specifies how a standard product is upgraded to premium.
// The plan is dynamic since it has to be calculated based on
// current valid renewal orders' balance.
// This plan is saved to `upgrade_plan` table.
// The data field is an array of each order's balance containing a
// upgrade memberID referring to the ID field. Each data element is saved
// to `proration` table.
type UpgradeSchema struct {
	ID        string            `db:"upgrade_id"`
	CreatedAt chrono.Time       `db:"created_at"`
	Balance   float64           `db:"balance"` // Deprecate: use Wallet.Balance instead.
	//Data      []ProratedOrderSchema `json:"data"` // Deprecate

	//Wallet    Wallet            `json:"wallet"`
	plan.Plan

}

// Deprecate
//func NewUpgradeIntent(wallet Wallet, p plan.Plan) UpgradeSchema {
//	return UpgradeSchema{
//		ID:        GenerateUpgradeID(),
//		Balance:   wallet.Balance,
//		CreatedAt: chrono.TimeNow(),
//		Data:      []ProratedOrderSchema{}, // This initializes to an empty array, rather than nil so that when marshaling into JSON, it is [], not null.
//		Wallet:    wallet,
//		Plan:      p.WithUpgrade(wallet.Balance),
//	}
//}

// Deprecate
//func (up UpgradeSchema) ProrationSources() []ProratedOrderSchema {
//	sources := make([]ProratedOrderSchema, 0)
//
//	for _, v := range up.Wallet.Source {
//		s := ProratedOrderSchema{
//			ProratedOrder: v,
//			CreatedUTC:    chrono.TimeNow(),
//			ConsumedUTC:   chrono.Time{},
//			UpgradeID:     up.ID,
//		}
//
//		sources = append(sources, s)
//	}
//
//	return sources
//}

// ReadableBalance produces a string describing the total balance
// in the format: CNY99.00 in email sent to user.
// Deprecate
func (up UpgradeSchema) ReadableBalance() string {
	return fmt.Sprintf("%s%.2f", "CNY", up.Balance)
}
