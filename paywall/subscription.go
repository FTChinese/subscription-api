package paywall

import (
	"fmt"
	"github.com/FTChinese/go-rest"
	"github.com/pkg/errors"
	"gitlab.com/ftchinese/subscription-api/util"
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

	return "FT" + strings.ToUpper(id), nil
}

type Charge struct {
	ListPrice   float64 `json:"listPrice"`
	NetPrice    float64 `json:"netPrice"` // Deprecate
	Amount      float64 `json:"amount"`
	IsConfirmed bool    `json:"-"`
}

// AliPrice converts Charged price to ailpay format
func (c Charge) AliPrice() string {
	return strconv.FormatFloat(c.Amount, 'f', 2, 32)
}

// PriceInCent converts Charged price to int64 in cent for comparison with wx notification.
func (c Charge) PriceInCent() int64 {
	return int64(c.Amount * 100)
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
	Charge
	ConfirmedAt chrono.Time `json:"-"` // When the payment is confirmed.
	Coordinate
	CreatedAt     chrono.Time    `json:"createdAt"`
	Currency      string         `json:"-"`
	CycleCount    int64          `json:"cycleCount"` // Default to 1. Change it for upgrade
	EndDate       chrono.Date    `json:"-"`          // Membership end date for this order. Depends on start date.
	ExtraDays     int64          `json:"extraDays"`  // Default to 1. Change it for upgraded.
	ID            string         `json:"id"`
	PaymentMethod enum.PayMethod `json:"payMethod"`
	StartDate     chrono.Date    `json:"-"`         // Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	Usage         SubsKind       `json:"usageType"` // The usage of this order: creat new, renew, or upgrade?
	AccountID
	WxAppID null.String `json:"-"` // Wechat specific
}

// NewSubs creates a new subscription with shared fields
// populated. PaymentMethod, Usage, UpgradeSource,
// UpgradeBalance are left to the controller layer.
func NewSubs(u AccountID, p Plan) (Subscription, error) {
	id, err := GenerateOrderID()

	if err != nil {
		return Subscription{}, err
	}

	s := Subscription{
		CreatedAt: chrono.TimeNow(),
		Charge: Charge{
			ListPrice: p.ListPrice,
			NetPrice:  p.NetPrice,
			Amount:    p.NetPrice,
		},
		Coordinate: Coordinate{
			Tier:  p.Tier,
			Cycle: p.Cycle,
		},
		Currency:   p.Currency,
		CycleCount: p.CycleCount,
		ExtraDays:  p.ExtraDays,
		ID:         id,
		AccountID:  u,
	}

	return s, nil
}

// NewUpgradeOrder creates an upgrade order.
// Deprecate
func NewUpgradeOrder(u AccountID, up Upgrade) (Subscription, error) {
	s := Subscription{
		Charge: Charge{
			ListPrice: up.ListPrice,
			NetPrice:  up.NetPrice,
			Amount:    up.NetPrice,
		},
		CreatedAt: chrono.TimeNow(),
		Coordinate: Coordinate{
			Tier:  up.Tier,
			Cycle: up.Cycle,
		},
		Currency:   up.Currency,
		CycleCount: up.CycleCount,
		ExtraDays:  up.ExtraDays,
		Usage:      SubsKindUpgrade,
		AccountID:  u,
	}

	id, err := GenerateOrderID()

	if err != nil {
		return s, err
	}

	s.ID = id

	return s, nil
}

func (s Subscription) ReadableAmount() string {
	return fmt.Sprintf("%s%.2f",
		strings.ToUpper(s.Currency),
		s.Amount,
	)
}

// Validate ensures the order to confirm must match
// the state of membership prior to creation/upgrading.
// If subs.Usage == SubsKindCreate, member.Tier == InvalidTier;
// If subs.Usage == SubsKindRenew, member.Tier == subs.Tier;
// If subs.Usage == SubsKindUpgrade, member.Tier != subs.Tier && member.Tier != TierInvalid
func (s Subscription) Validate(m Membership) error {
	switch s.Usage {

	case SubsKindUpgrade:
		if s.Tier != enum.TierPremium {
			return util.ErrTierMismatched
		}
		if m.Tier == enum.InvalidTier {
			return util.ErrNoUpgradingTarget
		}
		// For upgrading, order's tier must be different
		// from member's tier; otherwise this might be
		// a duplicate upgrading request.
		if m.Tier == s.Tier {
			return util.ErrDuplicateUpgrading
		}

	default:
		return nil
	}

	return nil
}

func (s Subscription) GetStartDate(m Membership, confirmedAt time.Time) time.Time {
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
		if s.Usage == SubsKindUpgrade {
			startTime = confirmedAt
		} else {
			startTime = m.ExpireDate.Time
		}
	}

	return startTime
}

func (s Subscription) GetEndDate(startTime time.Time) (time.Time, error) {
	var endTime time.Time

	switch s.Cycle {
	case enum.CycleYear:
		endTime = startTime.AddDate(int(s.CycleCount), 0, int(s.ExtraDays))

	case enum.CycleMonth:
		endTime = startTime.AddDate(0, int(s.CycleCount), int(s.ExtraDays))

	default:
		return endTime, errors.New("invalid billing cycle")
	}

	return endTime, nil
}

// WithMember updates an order with existing membership.
// Zero membership is a valid value.
func (s Subscription) Confirm(m Membership, confirmedAt time.Time) (Subscription, error) {

	startTime := s.GetStartDate(m, confirmedAt)
	endTime, err := s.GetEndDate(startTime)
	if err != nil {
		return s, err
	}

	s.ConfirmedAt = chrono.TimeFrom(confirmedAt)
	s.StartDate = chrono.DateFrom(startTime)
	s.EndDate = chrono.DateFrom(endTime)
	s.IsConfirmed = true

	return s, nil
}
