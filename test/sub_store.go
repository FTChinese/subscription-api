package test

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"time"
)

// BuildSubs generates Subscription for the following
// combination matrix:
// ftcOnlyId       wechatPay   create
// wechatOnlyId    aliPay      renew
// linkedId					   upgrade

// SubStore is a mock database for a single member.
// It mimics the working flow of user's order and membership
// creation and updating inside a real db.
type SubStore struct {
	Profile Profile
	Orders  map[string]subscription.Order // A user could have multiple orders.
	Member  subscription.Membership       // But only one membership.

	Snapshot subscription.MemberSnapshot // This will be populated and updated for any order other than `create`.

	balanceAnchor time.Time
	accountKind   reader.AccountKind
	payMethod     enum.PayMethod
	plan          plan.Plan
}

// NewSubStore creates a new storage for a user's membership.
func NewSubStore(p Profile) *SubStore {

	return &SubStore{
		Profile: p,
		Orders:  make(map[string]subscription.Order), // Initially user has no orders.
		Member:  subscription.Membership{},

		// Control behavior of orders and membership.
		accountKind:   reader.AccountKindFtc,
		balanceAnchor: time.Now(),
		payMethod:     RandomPayMethod(),
		plan:          YearlyStandard,
	}
}

func (s *SubStore) SetAccountKind(k reader.AccountKind) *SubStore {
	s.accountKind = k
	return s
}

func (s *SubStore) SetBalanceAnchor(t time.Time) *SubStore {
	s.balanceAnchor = t
	return s
}

func (s *SubStore) SetPlan(p plan.Plan) *SubStore {
	s.plan = p
	return s
}

func (s *SubStore) GetBalanceSource() []subscription.ProratedOrderSchema {
	sources := make([]subscription.ProratedOrderSchema, 0)

	for _, v := range s.Orders {
		if !v.IsConfirmed() {
			continue
		}

		if s.balanceAnchor.IsZero() {
			s.balanceAnchor = time.Now()
		}

		if v.Tier != enum.TierStandard {
			continue
		}

		if v.EndDate.Before(s.balanceAnchor) {
			continue
		}

		sources = append(sources, subscription.ProratedOrderSchema{})
	}

	return sources
}

func (s *SubStore) GetWallet() subscription.Wallet {
	orders := make([]subscription.ProratedOrder, 0)

	if s.balanceAnchor.IsZero() {
		s.balanceAnchor = time.Now()
	}

	for _, v := range s.Orders {
		if v.IsZero() || !v.IsConfirmed() {
			continue
		}

		if v.Tier != enum.TierStandard {
			continue
		}

		if v.EndDate.Time.Before(s.balanceAnchor) {
			continue
		}

		o := subscription.ProratedOrder{
			OrderID:   v.ID,
			Amount:    v.Amount,
			StartDate: v.StartDate,
			EndDate:   v.EndDate,
			Balance:   0,
		}

		orders = append(orders, o)
	}

	return subscription.NewWallet(orders, time.Now())
}

func (s *SubStore) MustCreateOrder() subscription.Order {

	builder := subscription.NewOrderBuilder(s.Profile.AccountID(s.accountKind)).
		SetPlan(s.plan).
		SetPayMethod(s.payMethod).
		SetMembership(s.Member).
		SetClient(RandomClientApp()).
		SetWallet(s.GetWallet())

	if s.payMethod == enum.PayMethodWx {
		builder.SetWxAppID(WxPayApp.AppID)
	}

	builder.MustBuild()

	order, _ := builder.Order()

	s.Orders[order.ID] = order

	return order
}

func (s *SubStore) MustConfirmOrder(id string) subscription.Order {
	o, err := s.getOrder(id)

	if err != nil {
		panic(err)
	}

	builder := subscription.
		NewConfirmationBuilder(subscription.PaymentResult{
			Amount:      o.AmountInCent(true),
			OrderID:     o.ID,
			ConfirmedAt: time.Now(),
		}, true).
		SetOrder(o).
		SetMembership(s.Member)

	if err := builder.ValidateOrder(); err != nil {
		panic(err)
	}

	builder.MustBuild()

	order := builder.ConfirmedOrder()
	member := builder.ConfirmedMembership()

	// Add the confirmed order back to store.
	s.Orders[o.ID] = order

	s.Member = member

	return order
}

func (s *SubStore) MustRenewN(n int) []subscription.Order {
	orders := make([]subscription.Order, 0)

	for i := 0; i < n; i++ {
		o := s.MustCreateOrder()

		o = s.MustConfirmOrder(o.ID)

		orders = append(orders, o)
	}

	return orders
}

func (s *SubStore) MustRenewalOrder() subscription.Order {
	order := s.MustCreateOrder()

	order = s.MustConfirmOrder(order.ID)

	order2 := s.MustCreateOrder()

	return order2
}

func (s *SubStore) MustUpgradingOrder() subscription.Order {
	order := s.MustCreateOrder()

	order = s.MustConfirmOrder(order.ID)

	order2 := s.MustCreateOrder()

	return order2
}

// getOrder retrieves a previously saved order.
func (s *SubStore) getOrder(id string) (subscription.Order, error) {
	o, ok := s.Orders[id]
	if !ok {
		return subscription.Order{}, fmt.Errorf("order %s is not found", id)
	}

	return o, nil
}
