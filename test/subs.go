package test

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/Pallinder/go-randomdata"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"time"
)

// BuildSubs generates Subscription for the following
// combination matrix:
// ftcOnlyId       wechatPay   create
// wechatOnlyId    aliPay      renew
// boundId					   upgrade
func BuildSubs(u paywall.UserID, pm enum.PayMethod, k paywall.SubsKind) paywall.Subscription {

	subs, err := paywall.NewSubs(
		u,
		YearlyStandard)

	if err != nil {
		panic(err)
	}

	subs.Usage = k

	switch pm {
	case enum.PayMethodWx:
		subs = subs.WithWxpay(WxPayClient.GetApp().AppID)
	case enum.PayMethodAli:
		subs = subs.WithAlipay()
	}

	return subs
}

// SubsRandom builds a random subscription order.
func SubsRandom(u paywall.UserID) paywall.Subscription {
	return BuildSubs(
		u,
		enum.PayMethod(randomdata.Number(1, 3)),
		paywall.SubsKind(randomdata.Number(1, 3)),
	)
}

// SubsCreate builds an order that is used to create a new
// member
func SubsCreate(u paywall.UserID) paywall.Subscription {
	return BuildSubs(
		u,
		enum.PayMethod(randomdata.Number(1, 3)),
		paywall.SubsKindCreate,
	)
}

// SubsRenew builds an order that is used to renew a member
func SubsRenew(u paywall.UserID) paywall.Subscription {
	return BuildSubs(
		u,
		enum.PayMethod(randomdata.Number(1, 3)),
		paywall.SubsKindRenew,
	)
}

// SubsUpgrade builds an order that is used to upgrade membership.
func SubsUpgrade(u paywall.UserID, up paywall.Upgrade) paywall.Subscription {

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

// SubsConfirmed builds an order that is confirmed.
func SubsConfirmed(u paywall.UserID) paywall.Subscription {
	subs := SubsRandom(u)

	subs, err := subs.Confirm(paywall.Membership{}, time.Now())

	if err != nil {
		panic(err)
	}

	return subs
}
