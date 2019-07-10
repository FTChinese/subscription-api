package model

import (
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/test"
	"time"
)

// helper function
func saveOrder(env Env, subs paywall.Subscription) {
	otx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	err = otx.SaveOrder(subs, test.RandomClientApp())
	if err != nil {
		panic(err)
	}

	if err := otx.commit(); err != nil {
		panic(err)
	}
}

func saveUpgradeOrder(env Env, subs paywall.Subscription) {
	otx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	err = otx.SaveOrder(subs, test.RandomClientApp())
	if err != nil {
		panic(err)
	}

	err = otx.SaveUpgradeSource(subs.OrderID, subs.UpgradeSource)
	if err != nil {
		panic(err)
	}

	if err := otx.commit(); err != nil {
		panic(err)
	}
}

func createMember(env Env, user paywall.UserID) {
	subs := test.SubsCreate(user)
	saveOrder(env, subs)

	mtx, err := env.BeginMemberTx()
	if err != nil {
		panic(err)
	}

	// Build a confirmed order.
	subs, err = subs.Confirm(paywall.Membership{}, time.Now())
	if err != nil {
		panic(err)
	}

	// Update the order
	err = mtx.ConfirmOrder(subs)
	if err != nil {
		panic(err)
	}

	// Build membership from order
	mm, err := subs.BuildMembership()
	if err != nil {
		panic(err)
	}

	// insert member.
	err = mtx.UpsertMember(mm)
	if err != nil {
		panic(err)
	}

	if err := mtx.commit(); err != nil {
		panic(err)
	}
}
