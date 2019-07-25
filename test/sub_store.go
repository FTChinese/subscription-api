package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"time"
)

// BuildSubs generates Subscription for the following
// combination matrix:
// ftcOnlyId       wechatPay   create
// wechatOnlyId    aliPay      renew
// boundId					   upgrade
func BuildSubs(
	accountID paywall.AccountID,
	pm enum.PayMethod,
	usage paywall.SubsKind,
) paywall.Subscription {

	id, err := paywall.GenerateOrderID()
	if err != nil {
		panic(err)
	}

	s := paywall.Subscription{
		Charge: paywall.Charge{
			ListPrice:   PlanStandardYearly.ListPrice,
			NetPrice:    PlanStandardYearly.NetPrice,
			IsConfirmed: false,
		},
		Coordinate: paywall.Coordinate{
			Tier:  PlanStandardYearly.Tier,
			Cycle: PlanStandardYearly.Cycle,
		},
		CreatedAt:     chrono.TimeNow(),
		Currency:      PlanStandardYearly.Currency,
		CycleCount:    PlanStandardYearly.CycleCount,
		ExtraDays:     PlanStandardYearly.ExtraDays,
		ID:            id,
		PaymentMethod: pm,
		Usage:         usage,
		User:          accountID,
	}

	if pm == enum.PayMethodWx {
		s.WxAppID = null.StringFrom(WxPayApp.AppID)
	}

	return s
}

// SubStore acts like a storage engine for a user's subscription.
type SubStore struct {
	User   paywall.AccountID
	Orders []paywall.Subscription // A user could have multiple orders.
	Member paywall.Membership     // But only one membership.
}

// NewSubStore creates a new storage for a user's membership.
func NewSubStore(id paywall.AccountID) *SubStore {
	return &SubStore{
		User:   id,
		Orders: make([]paywall.Subscription, 0), // Initially user has no orders.
		Member: paywall.Membership{},            // and zero membership.
	}
}

// AddOrder creates a new order and update user's membership.
func (s *SubStore) AddOrder(kind paywall.SubsKind) {
	// Build a new order
	o := BuildSubs(s.User, RandomPayMethod(), kind)

	// Confirm the order
	o, err := o.Confirm(s.Member, time.Now())
	if err != nil {
		panic(err)
	}

	// Store this order
	s.Orders = append(s.Orders, o)

	// Derive membership from this order.
	m, err := s.Member.FromAliOrWx(o)
	if err != nil {
		panic(err)
	}

	s.Member = m
}

// GetLastSub retrieves last order created.
func (s *SubStore) GetLastOrder() paywall.Subscription {
	l := len(s.Orders)

	if l > 0 {
		return s.Orders[l-1]
	}

	return paywall.Subscription{}
}
