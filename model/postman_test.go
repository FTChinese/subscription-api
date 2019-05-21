package model

import (
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
	"time"
)

// The prerequisite to test upgrading order in postman.
// Create two order's and confirm them.
// Create
func TestBuildUpgradeOrders(t *testing.T) {
	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	// You need to create at least one confirmed standard order.
	p := test.NewProfile()
	subsCreate := p.SubsCreate()
	subsRenew := p.SubsRenew()

	for _, subs := range []paywall.Subscription{subsCreate, subsRenew} {
		err := env.SaveSubscription(subs, test.RandomClientApp())
		if err != nil {
			panic(err)
		}

		_, err = env.ConfirmPayment(subs.OrderID, time.Now())
		if err != nil {
			panic(err)
		}
	}
}
