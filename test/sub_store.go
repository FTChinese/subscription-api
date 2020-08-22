package test

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/models/plan"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
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
// WARNING: SubStore is only used to mimic subscription
// paid via alipay or wxpay.
// For Stripe and Apple, they do not create orders on our side,
// so use Persona to generate mock data.
type SubStore struct {
	Profile *Persona
	Orders  map[string]subs.Order // A user could have multiple orders.
	Member  subs.Membership       // But only one membership.

	Snapshot subs.MemberSnapshot // This will be populated and updated for any order other than `create`.

	balanceAnchor time.Time
	accountKind   reader.AccountKind
	payMethod     enum.PayMethod
	plan          plan.Plan
}

// NewSubStore represents all the data stored for a single user.
func NewSubStore(p *Persona) *SubStore {

	return &SubStore{
		Profile: p,
		Orders:  make(map[string]subs.Order), // Initially user has no orders.
		Member:  subs.Membership{},

		// Control behavior of orders and membership.
		accountKind:   reader.AccountKindFtc,
		balanceAnchor: time.Now(),
		payMethod:     faker.RandomPayMethod(),
	}
}

func (s *SubStore) GetAccount() reader.Account {
	s.Profile.SetAccountKind(s.accountKind)

	return s.Profile.Account()
}

func (s *SubStore) SetAccountKind(k reader.AccountKind) *SubStore {
	s.accountKind = k
	return s
}

func (s *SubStore) GetMemberID() reader.MemberID {
	return s.Profile.AccountID()
}

func (s *SubStore) SetBalanceAnchor(t time.Time) *SubStore {
	s.balanceAnchor = t
	return s
}

func (s *SubStore) SetPlan(p plan.Plan) *SubStore {
	s.plan = p
	return s
}

func (s *SubStore) GetWallet() subs.Wallet {
	orders := make([]subs.BalanceSource, 0)

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

		o := subs.BalanceSource{
			OrderID:   v.ID,
			Amount:    v.Amount,
			StartDate: v.StartDate,
			EndDate:   v.EndDate,
		}

		orders = append(orders, o)
	}

	return subs.NewWallet(orders, time.Now())
}

func (s *SubStore) createOrderBuilder() *subs.OrderBuilder {
	builder := subs.NewOrderBuilder(s.Profile.AccountID())

	if s.payMethod == enum.PayMethodWx {
		builder.SetWxAppID(WxPayApp.AppID)
	}

	return builder
}

// MustSaveOrder creates a new order and save it in
// Orders array indexed by the order id.
//func (s *SubStore) MustCreateOrder() subs.Order {
//	builder := s.createOrderBuilder()
//
//	err := builder.Build()
//
//	if err != nil {
//		panic(err)
//	}
//
//	return order
//}

// MustConfirmOrder confirms and order, save the updated version,
// and then updated the membership based on this order.
func (s *SubStore) MustConfirmOrder(id string) subs.Order {
	o, err := s.getOrder(id)

	if err != nil {
		panic(err)
	}

	builder := subs.NewConfirmationBuilder(subs.PaymentResult{
		Amount:      o.AmountInCent(true),
		OrderID:     o.ID,
		ConfirmedAt: time.Now(),
	}, true).
		SetOrder(o).
		SetMembership(s.Member)

	if err := builder.ValidateOrder(); err != nil {
		panic(err)
	}

	confirmed, err := builder.Build()
	if err != nil {
		panic(err)
	}

	// Add the confirmed order back to store.
	s.Orders[o.ID] = confirmed.Order

	s.Member = confirmed.Membership

	return confirmed.Order
}

// MustGetMembership returns the current membership data.
// If the membership is zero, an order will be created,
// confirmed, and resulting membership updated.
//func (s *SubStore) MustGetMembership() subs.Membership {
//	if s.Member.IsZero() {
//		o := s.MustCreateOrder()
//		s.MustConfirmOrder(o.ID)
//	}
//
//	return s.Member
//}

//func (s *SubStore) MustRenewN(n int) []subs.Order {
//	orders := make([]subs.Order, 0)
//
//	for i := 0; i < n; i++ {
//		o := s.MustCreateOrder()
//
//		o = s.MustConfirmOrder(o.ID)
//
//		orders = append(orders, o)
//	}
//
//	return orders
//}

// MustUpgrade gets the ProratedOrderSchema.
// n specifies how many rounds you want the membership
// renewed before upgrading.
//func (s *SubStore) MustUpgrade(n int) (subs.UpgradeSchema, subs.Order) {
//	// Must have an valid membership before upgrading.
//	// Otherwise this is a new member.
//	if n < 1 {
//		panic("n must be greater than 1 to upgrade")
//	}
//	// Ensure the plan before upgrading is standard.
//	s.plan = YearlyStandard
//
//	s.MustRenewN(n)
//
//	// When changing the plan to premium, you are upgrading.
//	s.plan = YearlyPremium
//
//	builder := s.createOrderBuilder()
//
//	if err := builder.Build(); err != nil {
//		panic(err)
//	}
//
//	order, _ := builder.Order()
//
//	s.Orders[order.ID] = order
//
//	upgrade, err := builder.UpgradeSchema()
//	if err != nil {
//		panic(err)
//	}
//
//	return upgrade, order
//}

//func (s *SubStore) MustRenewalOrder() subs.Order {
//	order := s.MustCreateOrder()
//
//	order = s.MustConfirmOrder(order.ID)
//
//	order2 := s.MustCreateOrder()
//
//	return order2
//}

//func (s *SubStore) MustUpgradingOrder() subs.Order {
//	order := s.MustCreateOrder()
//
//	order = s.MustConfirmOrder(order.ID)
//
//	order2 := s.MustCreateOrder()
//
//	return order2
//}

// getOrder retrieves a previously saved order.
func (s *SubStore) getOrder(id string) (subs.Order, error) {
	o, ok := s.Orders[id]
	if !ok {
		return subs.Order{}, fmt.Errorf("order %s is not found", id)
	}

	return o, nil
}
