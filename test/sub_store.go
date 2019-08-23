package test

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
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
			Amount:      PlanStandardYearly.NetPrice,
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

	if usage == paywall.SubsKindUpgrade {
		s.ListPrice = PlanPremiumYearly.ListPrice
		s.NetPrice = PlanPremiumYearly.NetPrice
		s.Amount = PlanPremiumYearly.NetPrice
		s.Tier = PlanPremiumYearly.Tier
		s.Cycle = PlanPremiumYearly.Cycle
	}
	return s
}

// SubStore is a mock database for a single member.
// It mimics the working flow of user's order and membership
// creation and updating inside a real db.
type SubStore struct {
	User      paywall.AccountID
	Orders    map[string]paywall.Subscription // A user could have multiple orders.
	Member    paywall.Membership              // But only one membership.
	UpgradeV1 paywall.Upgrade
	UpgradeV2 paywall.UpgradePlan
}

// NewSubStore creates a new storage for a user's membership.
func NewSubStore(id paywall.AccountID) *SubStore {
	return &SubStore{
		User:   id,
		Orders: make(map[string]paywall.Subscription), // Initially user has no orders.
		Member: paywall.Membership{},                  // and zero membership.
	}
}

// AddOrder creates a new order and update user's membership.
func (s *SubStore) AddOrder(kind paywall.SubsKind) paywall.Subscription {
	// Build a new order
	o := BuildSubs(s.User, RandomPayMethod(), kind)
	s.Orders[o.ID] = o

	return o
}

// GetOrder retrieves a previously saved order.
func (s *SubStore) GetOrder(id string) (paywall.Subscription, error) {
	o, ok := s.Orders[id]
	if !ok {
		return paywall.Subscription{}, fmt.Errorf("order %s is not found", id)
	}

	return o, nil
}

func (s *SubStore) ConfirmOrder(id string) (paywall.Subscription, error) {
	o, err := s.GetOrder(id)
	if err != nil {
		return o, err
	}

	if o.IsConfirmed {
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

func (s *SubStore) UpgradeOrder(n int) (paywall.Subscription, error) {
	sources := make([]paywall.ProrationSource, 0)

	o := s.AddOrder(paywall.SubsKindCreate)
	o, err := s.ConfirmOrder(o.ID)
	if err != nil {
		return o, err
	}

	sources = append(sources, paywall.ProrationSource{
		OrderID:    o.ID,
		PaidAmount: o.Amount,
		StartDate:  o.StartDate,
		EndDate:    o.EndDate,
	})

	for i := 0; i < n; i++ {
		o := s.AddOrder(paywall.SubsKindRenew)
		o, err := s.ConfirmOrder(o.ID)
		if err != nil {
			return o, err
		}

		sources = append(sources, paywall.ProrationSource{
			OrderID:    o.ID,
			PaidAmount: o.Amount,
			StartDate:  o.StartDate,
			EndDate:    o.EndDate,
		})
	}

	s.UpgradeV1 = paywall.NewUpgrade(PlanPremiumYearly).SetBalance(sources).CalculatePayable()
	s.UpgradeV1.Member = s.Member

	s.UpgradeV2 = paywall.NewUpgradePreview(sources)
	s.UpgradeV2.Member = s.Member

	o = s.AddOrder(paywall.SubsKindUpgrade)

	return o, nil
}
