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
	// Saves upgrade schema.
	upgrade map[string]subscription.UpgradeSchema
}

// NewSubStore represents all the data stored for a single user.
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
		upgrade:       make(map[string]subscription.UpgradeSchema),
	}
}

func (s *SubStore) GetAccount() reader.Account {
	return s.Profile.Account(s.accountKind)
}

func (s *SubStore) SetAccountKind(k reader.AccountKind) *SubStore {
	s.accountKind = k
	return s
}

func (s *SubStore) GetMemberID() reader.MemberID {
	return s.Profile.AccountID(s.accountKind)
}

func (s *SubStore) SetBalanceAnchor(t time.Time) *SubStore {
	s.balanceAnchor = t
	return s
}

func (s *SubStore) SetPlan(p plan.Plan) *SubStore {
	s.plan = p
	return s
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

func (s *SubStore) createOrderBuilder() *subscription.OrderBuilder {
	builder := subscription.NewOrderBuilder(s.Profile.AccountID(s.accountKind)).
		SetPlan(s.plan).
		SetPayMethod(s.payMethod).
		SetMembership(s.Member).
		SetClient(RandomClientApp()).
		SetWallet(s.GetWallet())

	if s.payMethod == enum.PayMethodWx {
		builder.SetWxAppID(WxPayApp.AppID)
	}

	return builder
}

// MustCreateOrder creates a new order and save it in
// Orders array indexed by the order id.
func (s *SubStore) MustCreateOrder() subscription.Order {
	builder := s.createOrderBuilder()

	err := builder.Build()

	if err != nil {
		panic(err)
	}

	order, _ := builder.Order()

	s.Orders[order.ID] = order

	return order
}

// MustConfirmOrder confirms and order, save the updated version,
// and then updated the membership based on this order.
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

// MustGetMembership returns the current membership data.
// If the membership is zero, an order will be created,
// confirmed, and resulting membership updated.
func (s *SubStore) MustGetMembership() subscription.Membership {
	if s.Member.IsZero() {
		o := s.MustCreateOrder()
		s.MustConfirmOrder(o.ID)
	}

	return s.Member
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

// MustUpgrade gets the ProratedOrderSchema.
// n specifies how many rounds you want the membership
// renewed before upgrading.
func (s *SubStore) MustUpgrade(n int) (subscription.UpgradeSchema, subscription.Order) {
	// Must have an valid membership before upgrading.
	// Otherwise this is a new member.
	if n < 1 {
		panic("n must be greater than 1 to upgrade")
	}
	// Ensure the plan before upgrading is standard.
	s.plan = YearlyStandard

	s.MustRenewN(n)

	// When changing the plan to premium, you are upgrading.
	s.plan = YearlyPremium

	builder := s.createOrderBuilder()

	if err := builder.Build(); err != nil {
		panic(err)
	}

	order, _ := builder.Order()

	s.Orders[order.ID] = order

	upgrade, err := builder.UpgradeSchema()
	if err != nil {
		panic(err)
	}

	return upgrade, order
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
