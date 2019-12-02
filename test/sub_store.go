package test

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
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
	Profile       Profile
	AccountID     reader.MemberID
	Orders        map[string]subscription.Order // A user could have multiple orders.
	Member        subscription.Membership       // But only one membership.
	balanceAnchor time.Time
	UpgradePlan   subscription.UpgradeIntent  // To have data populated, you must call MustRenewN() and then call MustCreate(PremiumPlan).
	Snapshot      subscription.MemberSnapshot // This will be populated and updated for any order other than `create`.
}

// NewSubStore creates a new storage for a user's membership.
func NewSubStore(p Profile, k reader.AccountKind) *SubStore {

	return &SubStore{
		Profile:       p,
		AccountID:     p.AccountID(k),
		Orders:        make(map[string]subscription.Order), // Initially user has no orders.
		Member:        subscription.Membership{},           // and zero membership.
		balanceAnchor: time.Now(),
	}
}

func (s *SubStore) Backdate(n int) *SubStore {
	s.balanceAnchor = s.balanceAnchor.AddDate(0, 0, n)

	return s
}

func (s *SubStore) CreateOrder(p plan.Plan) (subscription.Order, error) {
	kind, _ := s.Member.SubsKind(p)
	order, err := subscription.NewOrder(
		s.AccountID,
		p,
		RandomPayMethod(),
		kind)

	if err != nil {
		return order, err
	}

	if order.Usage == subscription.SubsKindUpgrade {
		sources := s.GetBalanceSource()

		up := subscription.NewUpgradeIntent(sources)

		order = order.WithUpgrade(up)

		s.UpgradePlan = up
	}

	if order.PaymentMethod == enum.PayMethodWx {
		order.WxAppID = null.StringFrom(WxPayApp.AppID)
	}

	if !s.Member.IsZero() {
		snapshot := subscription.NewMemberSnapshot(s.Member, order.Usage.SnapshotReason())
		order.MemberSnapshotID = null.StringFrom(snapshot.SnapshotID)

		s.Snapshot = snapshot
	}

	s.Orders[order.ID] = order

	return order, nil
}

func (s *SubStore) MustCreate(p plan.Plan) subscription.Order {
	order, err := s.CreateOrder(p)
	if err != nil {
		panic(err)
	}

	return order
}

// ConfirmOrder confirms an existing order
func (s *SubStore) ConfirmOrder(id string) (subscription.Order, error) {
	o, err := s.GetOrder(id)
	if err != nil {
		return o, err
	}

	if o.IsConfirmed() {
		return o, fmt.Errorf("order %s is already confirmed", id)
	}

	o, err = o.Confirm(s.Member, time.Now())
	if err != nil {
		return o, err
	}

	// Add the confirmed order back to store.
	s.Orders[o.ID] = o

	m, err := s.Member.FromAliOrWx(o)
	if err != nil {
		return o, err
	}

	s.Member = m

	return o, nil
}

func (s *SubStore) MustConfirm(id string) subscription.Order {
	order, err := s.ConfirmOrder(id)

	if err != nil {
		panic(err)
	}

	return order
}

// RenewN creates a new membership and renew it multiple times.
//
func (s *SubStore) RenewN(p plan.Plan, n int) ([]subscription.Order, error) {

	orders := []subscription.Order{}

	for i := 0; i < n; i++ {
		o, err := s.CreateOrder(p)
		if err != nil {
			return nil, err
		}

		o, err = s.ConfirmOrder(o.ID)
		if err != nil {
			return nil, err
		}

		orders = append(orders, o)
	}

	return orders, nil
}

func (s *SubStore) MustRenewN(p plan.Plan, n int) []subscription.Order {
	orders, err := s.RenewN(p, n)
	if err != nil {
		panic(err)
	}

	return orders
}

// RenewalOrder creates an order used for renewal
func (s *SubStore) RenewalOrder(p plan.Plan) (subscription.Order, error) {
	order, err := s.CreateOrder(p)
	if err != nil {
		return subscription.Order{}, err
	}

	_, err = s.ConfirmOrder(order.ID)
	if err != nil {
		return subscription.Order{}, err
	}

	order2, err := s.CreateOrder(p)
	if err != nil {
		return subscription.Order{}, err
	}

	return order2, nil
}

func (s *SubStore) MustRenewal(p plan.Plan) subscription.Order {
	order, err := s.RenewalOrder(p)
	if err != nil {
		panic(err)
	}

	return order
}

// UpgradingOrder creates an order used for upgrading.
func (s *SubStore) UpgradingOrder() (subscription.Order, error) {
	order, err := s.CreateOrder(YearlyStandard)
	if err != nil {
		return subscription.Order{}, err
	}

	_, err = s.ConfirmOrder(order.ID)
	if err != nil {
		return subscription.Order{}, err
	}

	order2, err := s.CreateOrder(YearlyPremium)
	if err != nil {
		return subscription.Order{}, err
	}

	return order2, nil
}

func (s *SubStore) MustUpgrading() subscription.Order {
	order, err := s.UpgradingOrder()
	if err != nil {
		panic(err)
	}

	return order
}

// GetOrder retrieves a previously saved order.
func (s *SubStore) GetOrder(id string) (subscription.Order, error) {
	o, ok := s.Orders[id]
	if !ok {
		return subscription.Order{}, fmt.Errorf("order %s is not found", id)
	}

	return o, nil
}

func (s *SubStore) GetBalanceSource() []subscription.ProrationSource {
	sources := []subscription.ProrationSource{}

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

		sources = append(sources, subscription.ProrationSource{
			OrderID:    v.ID,
			PaidAmount: v.Amount,
			StartDate:  v.StartDate,
			EndDate:    v.EndDate,
			//Balance:     0,
			//CreatedUTC:  chrono.Time{},
			//ConsumedUTC: chrono.Time{},
			//UpgradeID:   "",
		})
	}

	return sources
}
