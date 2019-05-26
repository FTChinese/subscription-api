package test

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/Pallinder/go-randomdata"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"time"
)

// BuildSubs generates Subscription for the following
// combination matrix:
// ftcOnlyId       wechatPay   create
// wechatOnlyId    aliPay      renew
// boundId					   upgrade
func BuildSubs(u paywall.User, pm enum.PayMethod, k paywall.SubsKind) paywall.Subscription {

	var subs paywall.Subscription
	var err error

	if k == paywall.SubsKindUpgrade {
		subs, err = paywall.NewSubsUpgrade(
			u,
			GenUpgradePlan())

		if err != nil {
			panic(err)
		}
	} else {
		subs, err = paywall.NewSubs(
			u,
			YearlyStandard)

		if err != nil {
			panic(err)
		}

		subs.Kind = k
	}

	switch pm {
	case enum.PayMethodWx:
		subs = subs.WithWxpay(WxPayClient.GetApp().AppID)
	case enum.PayMethodAli:
		subs = subs.WithAlipay()
	}

	return subs
}

// SubsRandom builds a random subscription order.
func SubsRandom(u paywall.User) paywall.Subscription {
	return BuildSubs(
		u,
		enum.PayMethod(randomdata.Number(1, 3)),
		paywall.SubsKind(randomdata.Number(1, 3)),
	)
}

// SubsCreate builds an order that is used to create a new
// member
func SubsCreate(u paywall.User) paywall.Subscription {
	return BuildSubs(
		u,
		enum.PayMethod(randomdata.Number(1, 3)),
		paywall.SubsKindCreate,
	)
}

// SubsRenew builds an order that is used to renew a member
func SubsRenew(u paywall.User) paywall.Subscription {
	return BuildSubs(
		u,
		enum.PayMethod(randomdata.Number(1, 3)),
		paywall.SubsKindRenew,
	)
}

// SubsUpgrade builds an order that is used to upgrade membership.
func SubsUpgrade(u paywall.User) paywall.Subscription {
	return BuildSubs(
		u,
		enum.PayMethod(randomdata.Number(1, 3)),
		paywall.SubsKindUpgrade,
	)
}

// SubsConfirmed builds an order that is confirmed.
func SubsConfirmed(u paywall.User) paywall.Subscription {
	subs := SubsRandom(u)

	subs, err := subs.ConfirmWithMember(paywall.Membership{}, time.Now())

	if err != nil {
		panic(err)
	}

	return subs
}
