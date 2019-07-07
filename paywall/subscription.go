package paywall

import (
	"github.com/FTChinese/go-rest"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

// GenerateOrderID creates an order id.
// The id has a total length of 18 chars.
// If we use this generator:
// `FT` takes 2, followed by year-month-date-hour-minute
// FT201905191139, then only 4 chars left for random number
// 2^16 = 65536, which means only 60000 order could be created every minute.
// To leave enough space for random number, 8 chars might be reasonable - 22 chars totally.
// If we use current random number generator:
// 2 ^ 64 = 1.8 * 10^19 orders.
func GenerateOrderID() (string, error) {

	id, err := gorest.RandomHex(8)
	if err != nil {
		return "", err
	}

	time.Now().Format("200601021504")
	return "FT" + strings.ToUpper(id), nil
}

type Charge struct {
	ListPrice   float64 `json:"listPrice"`
	NetPrice    float64 `json:"netPrice"`
	IsConfirmed bool    `json:"-"`
}

// AliNetPrice converts Charged price to ailpay format
func (c Charge) AliNetPrice() string {
	return strconv.FormatFloat(c.NetPrice, 'f', 2, 32)
}

// PriceInCent converts Charged price to int64 in cent for comparison with wx notification.
func (c Charge) PriceInCent() int64 {
	return int64(c.NetPrice * 100)
}

// Subscription contains the details of a user's action to place an order.
// This is the centrum of the whole subscription process.
// An order could represents 12 status of user:
// A user is allowed to to at max 2 ids - ftc or wechat, or both. This is 3 possible choices.;
// A user could choose between 2 payment methods;
// An order could create, renew or upgrade a member.
// And tier + cycle have 3 combination.
// All those combination add up to 3 * 2 * 3 * 3 = 54
type Subscription struct {
	// Fields common to all.
	OrderID string `json:"id"`
	UserID
	Charge
	TierToBuy    enum.Tier  `json:"tier"`
	BillingCycle enum.Cycle `json:"cycle"`
	CycleCount   int64      `json:"cycleCount"` // Default to 1. Change it for upgrade
	ExtraDays    int64      `json:"extraDays"`  // Default to 1. Change it for upgraded.

	// The category of this order.
	Kind SubsKind `json:"usageType"`

	// Fields only applicable to upgrade.
	UpgradeSource  []string   `json:"-"`       // for upgrade
	UpgradeBalance null.Float `json:"balance"` // for upgrade

	// Payment method
	PaymentMethod enum.PayMethod `json:"paymentMethod"`
	WxAppID       null.String    `json:"-"` // Wechat specific

	CreatedAt chrono.Time `json:"createdAt"`

	// Fields populated only after payment finished.
	ConfirmedAt chrono.Time `json:"confirmedAt"` // When the payment is confirmed.
	StartDate   chrono.Date `json:"startDate"`   // Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	EndDate     chrono.Date `json:"endDate"`     // Membership end date for this order. Depends on start date.
}

// NewSubs creates a new subscription with shared fields
// populated. PaymentMethod, Kind, UpgradeSource,
// UpgradeBalance are left to the controller layer.
func NewSubs(u UserID, p Plan) (Subscription, error) {
	s := Subscription{
		TierToBuy:    p.Tier,
		BillingCycle: p.Cycle,
		CycleCount:   1,
		ExtraDays:    1,
		CreatedAt:    chrono.TimeNow(),
	}

	s.CompoundID = u.CompoundID
	s.FtcID = u.FtcID
	s.UnionID = u.UnionID
	s.ListPrice = p.ListPrice
	s.NetPrice = p.NetPrice

	id, err := GenerateOrderID()

	if err != nil {
		return s, err
	}

	s.OrderID = id

	return s, nil
}

// NewSubsUpgrade creates an upgrade order.
func NewSubsUpgrade(u UserID, p UpgradePlan) (Subscription, error) {
	s := Subscription{
		TierToBuy:     p.Tier,
		BillingCycle:  p.Cycle,
		CycleCount:    p.CycleCount,
		ExtraDays:     p.ExtraDays,
		Kind:          SubsKindUpgrade,
		UpgradeSource: p.OrderIDs,
		// This value should always exist even for 0.
		UpgradeBalance: null.FloatFrom(p.Balance),
	}

	s.CompoundID = u.CompoundID
	s.FtcID = u.FtcID
	s.UnionID = u.UnionID
	s.ListPrice = p.ListPrice
	s.NetPrice = p.Payable

	id, err := GenerateOrderID()

	if err != nil {
		return s, err
	}

	s.OrderID = id

	return s, nil
}

// WithWxpay sets payment method to wechat
func (s Subscription) WithWxpay(appID string) Subscription {
	s.PaymentMethod = enum.PayMethodWx
	s.WxAppID = null.StringFrom(appID)

	return s
}

// WithAlipay sets payment method to alipay
func (s Subscription) WithAlipay() Subscription {
	s.PaymentMethod = enum.PayMethodAli

	return s
}

// UpgradeSourceIDs is used to render templates.
func (s Subscription) UpgradeSourceIDs() string {
	if s.UpgradeSource == nil {
		return ""
	}

	return strings.Join(s.UpgradeSource, ",")
}

// IsNewMember checks whether this order is used to create a
// new member
func (s Subscription) IsNewMember() bool {
	return s.Kind == SubsKindCreate
}

// IsRenewal checks whether this order is used to renew
// membership.
func (s Subscription) IsRenewal() bool {
	return s.Kind == SubsKindRenew
}

// IsUpgrade checks whether this order is used to upgrade
// member tier.
func (s Subscription) IsUpgrade() bool {
	return s.Kind == SubsKindUpgrade
}

// IsValidPay checks whether user actually used any payment
// method.
func (s Subscription) IsValidPay() bool {
	return s.PaymentMethod != enum.InvalidPay
}

// SetDuration updates the StartDate and EndDate fields.
func (s *Subscription) SetDuration(start time.Time) error {

	var endTime time.Time

	switch s.BillingCycle {
	case enum.CycleYear:
		endTime = start.AddDate(int(s.CycleCount), 0, int(s.ExtraDays))

	case enum.CycleMonth:
		endTime = start.AddDate(0, int(s.CycleCount), int(s.ExtraDays))

	default:
		return errors.New("invalid billing cycle")
	}

	s.StartDate = chrono.DateFrom(start)
	s.EndDate = chrono.DateFrom(endTime)

	return nil
}

// Validate ensures the order to confirm must match
// the state of membership prior to creation/upgrading.
// If subs.Kind == SubsKindCreate, member.Tier == InvalidTier;
// If subs.Kind == SubsKindRenew, member.Tier == subs.Tier;
// If subs.Kind == SubsKindUpgrade, member.Tier != subs.Tier && member.Tier != TierInvalid
func (s Subscription) Validate(m Membership) error {
	switch s.Kind {

	case SubsKindUpgrade:
		if s.TierToBuy != enum.TierPremium {
			return ErrTierMismatched
		}
		if m.Tier == enum.InvalidTier {
			return ErrNoUpgradingTarget
		}
		// For upgrading, order's tier must be different
		// from member's tier; otherwise this might be
		// a duplicate upgrading request.
		if m.Tier == s.TierToBuy {
			return ErrDuplicateUpgrading
		}

	default:
		return nil
	}

	return nil
}

// WithMember updates an order with existing membership.
// Zero membership is a valid value.
func (s Subscription) ConfirmWithMember(m Membership, confirmedAt time.Time) (Subscription, error) {
	s.ConfirmedAt = chrono.TimeFrom(confirmedAt)

	var startTime time.Time

	// If membership is expired, always use the confirmation
	// time as start time.
	if m.IsExpired() {
		startTime = confirmedAt
	} else {
		// If membership is not expired, this order might be
		// used to either renew or upgrade.
		// For renewal, we use current membership's
		// expiration date;
		// For upgrade, we use confirmation time.
		if s.Kind == SubsKindUpgrade {
			startTime = confirmedAt
		} else {
			startTime = m.ExpireDate.Time
		}
	}

	err := s.SetDuration(startTime)

	if err != nil {
		return s, err
	}

	s.IsConfirmed = true
	return s, nil
}

// BuildMembership creates a Membership based on a
// confirmed order.
func (s Subscription) BuildMembership() (Membership, error) {
	if !s.IsConfirmed {
		return Membership{}, errors.New("only confirmed order could be used to build membership")
	}

	return Membership{
		CompoundID: s.CompoundID,
		FTCUserID:  s.FtcID,
		UnionID:    s.UnionID,
		Tier:       s.TierToBuy,
		Cycle:      s.BillingCycle,
		ExpireDate: s.EndDate,
	}, nil
}
