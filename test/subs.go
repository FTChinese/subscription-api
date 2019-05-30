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

	subs, err := paywall.NewSubs(
		u,
		YearlyStandard)

	if err != nil {
		panic(err)
	}

	subs.Kind = k

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
func SubsUpgrade(u paywall.User, orders []paywall.Subscription) paywall.Subscription {
	unused := make([]paywall.UnusedOrder, 0)
	for _, v := range orders {
		unused = append(unused, paywall.UnusedOrder{
			ID:        v.OrderID,
			NetPrice:  v.NetPrice,
			StartDate: v.StartDate,
			EndDate:   v.EndDate,
		})
	}

	up := paywall.NewUpgradePlan(YearlyPremium).
		SetBalance(unused).
		CalculatePayable()

	subs, err := paywall.NewSubsUpgrade(u, up)

	if err != nil {
		panic(err)
	}

	pm := enum.PayMethod(randomdata.Number(1, 3))

	switch pm {
	case enum.PayMethodWx:
		subs = subs.WithWxpay(WxPayClient.GetApp().AppID)
	case enum.PayMethodAli:
		subs = subs.WithAlipay()
	}

	return subs
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
