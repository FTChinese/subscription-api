package test

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/reader"
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
	AccountID     reader.AccountID
	Orders        map[string]paywall.Order // A user could have multiple orders.
	Member        paywall.Membership       // But only one membership.
	balanceAnchor time.Time
	UpgradePlan   paywall.UpgradePlan    // To have data populated, you must call MustRenewN() and then call MustCreate(PremiumPlan).
	Snapshot      paywall.MemberSnapshot // This will be populated and updated for any order other than `create`.
}

// NewSubStore creates a new storage for a user's membership.
func NewSubStore(p Profile, k reader.AccountKind) *SubStore {

	return &SubStore{
		Profile:       p,
		AccountID:     p.AccountID(k),
		Orders:        make(map[string]paywall.Order), // Initially user has no orders.
		Member:        paywall.Membership{},           // and zero membership.
		balanceAnchor: time.Now(),
	}
}

func (s *SubStore) Backdate(n int) *SubStore {
	s.balanceAnchor = s.balanceAnchor.AddDate(0, 0, n)

	return s
}

func (s *SubStore) CreateOrder(p paywall.Plan) (paywall.Order, error) {
	order, err := paywall.NewOrder(
		s.AccountID,
		p,
		RandomPayMethod(),
		s.Member)

	if err != nil {
		return order, err
	}

	if order.Usage == paywall.SubsKindUpgrade {
		sources := s.GetBalanceSource()

		up := paywall.NewUpgradePlan(sources)

		order = order.WithUpgrade(up)

		s.UpgradePlan = up
	}

	if order.PaymentMethod == enum.PayMethodWx {
		order.WxAppID = null.StringFrom(WxPayApp.AppID)
	}

	if !s.Member.IsZero() {
		snapshot := paywall.NewMemberSnapshot(s.Member, order.Usage)
		order.MemberSnapshotID = null.StringFrom(snapshot.ID)

		s.Snapshot = snapshot
	}

	s.Orders[order.ID] = order

	return order, nil
}

func (s *SubStore) MustCreate(p paywall.Plan) paywall.Order {
	order, err := s.CreateOrder(p)
	if err != nil {
		panic(err)
	}

	return order
}

// ConfirmOrder confirms an existing order
func (s *SubStore) ConfirmOrder(id string) (paywall.Order, error) {
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

func (s *SubStore) MustConfirm(id string) paywall.Order {
	order, err := s.ConfirmOrder(id)

	if err != nil {
		panic(err)
	}

	return order
}

// RenewN creates a new membership and renew it multiple times.
//
func (s *SubStore) RenewN(p paywall.Plan, n int) ([]paywall.Order, error) {

	orders := []paywall.Order{}

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

func (s *SubStore) MustRenewN(p paywall.Plan, n int) []paywall.Order {
	orders, err := s.RenewN(p, n)
	if err != nil {
		panic(err)
	}

	return orders
}

// RenewalOrder creates an order used for renewal
func (s *SubStore) RenewalOrder(p paywall.Plan) (paywall.Order, error) {
	order, err := s.CreateOrder(p)
	if err != nil {
		return paywall.Order{}, err
	}

	_, err = s.ConfirmOrder(order.ID)
	if err != nil {
		return paywall.Order{}, err
	}

	order2, err := s.CreateOrder(p)
	if err != nil {
		return paywall.Order{}, err
	}

	return order2, nil
}

func (s *SubStore) MustRenewal(p paywall.Plan) paywall.Order {
	order, err := s.RenewalOrder(p)
	if err != nil {
		panic(err)
	}

	return order
}

// UpgradingOrder creates an order used for upgrading.
func (s *SubStore) UpgradingOrder() (paywall.Order, error) {
	order, err := s.CreateOrder(YearlyStandard)
	if err != nil {
		return paywall.Order{}, err
	}

	_, err = s.ConfirmOrder(order.ID)
	if err != nil {
		return paywall.Order{}, err
	}

	order2, err := s.CreateOrder(YearlyPremium)
	if err != nil {
		return paywall.Order{}, err
	}

	return order2, nil
}

func (s *SubStore) MustUpgrading() paywall.Order {
	order, err := s.UpgradingOrder()
	if err != nil {
		panic(err)
	}

	return order
}

// GetOrder retrieves a previously saved order.
func (s *SubStore) GetOrder(id string) (paywall.Order, error) {
	o, ok := s.Orders[id]
	if !ok {
		return paywall.Order{}, fmt.Errorf("order %s is not found", id)
	}

	return o, nil
}

func (s *SubStore) GetBalanceSource() []paywall.ProrationSource {
	sources := []paywall.ProrationSource{}

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

		sources = append(sources, paywall.ProrationSource{
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
