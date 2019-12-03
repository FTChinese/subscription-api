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

// UpgradeIntent specifies how a standard product is upgraded to premium.
// The plan is dynamic since it has to be calculated based on
// current valid renewal orders' balance.
// This plan is saved to `upgrade_plan` table.
// The data field is an array of each order's balance containing a
// upgrade id referring to the ID field. Each data element is saved
// to `proration` table.
type UpgradeIntent struct {
	ID        string            `json:"id" db:"upgrade_id"`
	Balance   float64           `json:"balance" db:"balance"` // Deprecate: use Wallet.Balance instead.
	CreatedAt chrono.Time       `json:"createdAt" db:"created_at"`
	Data      []ProrationSource `json:"data"` // Deprecate
	Wallet    Wallet            `json:"wallet"`
	Plan      plan.Plan         `json:"plan"`
}

func NewUpgradeIntent(wallet Wallet, p plan.Plan) UpgradeIntent {
	return UpgradeIntent{
		ID:        GenerateUpgradeID(),
		Balance:   wallet.Balance,
		CreatedAt: chrono.TimeNow(),
		Data:      []ProrationSource{}, // This initializes to an empty array, rather than nil so that when marshaling into JSON, it is [], not null.
		Wallet:    wallet,
		Plan:      p.WithUpgrade(wallet.Balance),
	}
}

func (up UpgradeIntent) ProrationSources() []ProrationSource {
	sources := make([]ProrationSource, 0)

	for _, v := range up.Wallet.Source {
		s := ProrationSource{
			ProratedOrder: v,
			CreatedUTC:    chrono.TimeNow(),
			ConsumedUTC:   chrono.Time{},
			UpgradeID:     up.ID,
		}

		sources = append(sources, s)
	}

	return sources
}

// ReadableBalance produces a string describing the total balance
// in the format: CNY99.00 in email sent to user.
// Deprecate
func (up UpgradeIntent) ReadableBalance() string {
	return fmt.Sprintf("%s%.2f", "CNY", up.Balance)
}
