package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/Pallinder/go-randomdata"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

// SubsRandom builds a random subscription order.
func SubsRandom(u paywall.AccountID) paywall.Subscription {
	return BuildSubs(
		u,
		enum.PayMethod(randomdata.Number(1, 3)),
		paywall.SubsKind(randomdata.Number(1, 3)),
	)
}

// SubsCreate builds an order that is used to create a new
// member
func SubsCreate(u paywall.AccountID) paywall.Subscription {
	return BuildSubs(
		u,
		enum.PayMethod(randomdata.Number(1, 3)),
		paywall.SubsKindCreate,
	)
}

// SubsRenew builds an order that is used to renew a member
func SubsRenew(u paywall.AccountID) paywall.Subscription {
	return BuildSubs(
		u,
		enum.PayMethod(randomdata.Number(1, 3)),
		paywall.SubsKindRenew,
	)
}

// SubsUpgrade builds an order that is used to upgrade membership.
func SubsUpgrade(u paywall.AccountID, up paywall.Upgrade) paywall.Subscription {

	subs, err := paywall.NewUpgradeOrder(u, up)

	if err != nil {
		panic(err)
	}

	pm := enum.PayMethod(randomdata.Number(1, 3))

	switch pm {
	case enum.PayMethodWx:
		subs.PaymentMethod = enum.PayMethodWx
		subs.WxAppID = null.StringFrom(WxPayApp.AppID)
	case enum.PayMethodAli:
		subs.PaymentMethod = enum.PayMethodAli
	}

	return subs
}
