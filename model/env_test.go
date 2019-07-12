package model

import (
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/test"
)

// Create a new order in db and returns it.
func createOrder(userID paywall.UserID) paywall.Subscription {
	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	order := test.SubsCreate(userID)

	if err := tx.SaveOrder(order, test.RandomClientApp()); err != nil {
		panic(err)
	}

	if err := tx.commit(); err != nil {
		panic(err)
	}

	return order
}

func createUpgrade(userID paywall.UserID) paywall.Upgrade {
	upgrade := test.GenUpgrade(userID)
	orderID, _ := paywall.GenerateOrderID()
	upgrade.Member = test.GenMember(userID, false)

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	if err := tx.SaveUpgrade(orderID, upgrade); err != nil {
		panic(err)
	}

	if err := tx.commit(); err != nil {
		panic(err)
	}

	return upgrade
}

func createMember(userID paywall.UserID) paywall.Membership {
	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	tx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	m := test.GenMember(userID, false)

	err = tx.CreateMember(m, null.String{})
	if err != nil {
		panic(err)
	}

	if err := tx.commit(); err != nil {
		panic(err)
	}

	return m
}
