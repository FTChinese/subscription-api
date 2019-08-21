package paywall

import (
	"fmt"
	"github.com/FTChinese/go-rest"
	"github.com/pkg/errors"
	util2 "gitlab.com/ftchinese/subscription-api/models/util"
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
	ListPrice   float64 `json:"listPrice" db:"price"`
	NetPrice    float64 `json:"netPrice" db:"price"` // Deprecate
	Amount      float64 `json:"amount" db:"amount"`
	IsConfirmed bool    `json:"-" db:"is_confirmed"`
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
	ID string `json:"id" db:"order_id"`
	AccountID
	//Charge
	ListPrice float64 `json:"listPrice" db:"price"`
	Amount    float64 `json:"amount" db:"amount"`
	Coordinate
	Currency      string         `json:"-"`
	CycleCount    int64          `json:"cycleCount" db:"cycle_count"` // Default to 1. Change it for upgrade
	ExtraDays     int64          `json:"extraDays" db:"extra_days"`   // Default to 1. Change it for upgraded.
	Usage         SubsKind       `json:"usageType" db:"usage_type"`   // The usage of this order: creat new, renew, or upgrade?
	LastUpgradeID null.String    `json:"-" db:"last_upgrade_id"`
	PaymentMethod enum.PayMethod `json:"payMethod" db:"payment_method"`
	WxAppID       null.String    `json:"-" db:"wx_app_id"`  // Wechat specific
	StartDate     chrono.Date    `json:"-" db:"start_date"` // Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	EndDate       chrono.Date    `json:"-" db:"end_date"`   // Membership end date for this order. Depends on start date.
	//User          AccountID      `json:"-"`                 // Deprecate
	CreatedAt   chrono.Time `json:"createdAt" db:"created_at"`
	ConfirmedAt chrono.Time `json:"-" db:"confirmed_at"` // When the payment is confirmed.
}

// NewSubs creates a new subscription with shared fields
// populated. PaymentMethod, Usage, UpgradeSource,
// UpgradeBalance are left to the controller layer.
func NewSubs(accountID AccountID, p Plan) (Subscription, error) {
	id, err := GenerateOrderID()

	if err != nil {
		return Subscription{}, err
	}

	s := Subscription{
		ID:        id,
		AccountID: accountID,
		ListPrice: p.ListPrice,
		Amount:    p.NetPrice,
		Coordinate: Coordinate{
			Tier:  p.Tier,
			Cycle: p.Cycle,
		},
		Currency:   p.Currency,
		CycleCount: p.CycleCount,
		ExtraDays:  p.ExtraDays,
		CreatedAt:  chrono.TimeNow(),
	}

	return s, nil
}

// NewOrder creates a new subscription order.
// If later it is found that this order is used for upgrading,
// upgrade it and returns a new instance with upgrading price.
func NewOrder(
	id AccountID,
	p Plan,
	method enum.PayMethod,
	m Membership,
) (Subscription, error) {
	orderID, err := GenerateOrderID()

	if err != nil {
		return Subscription{}, err
	}

	kind, err := m.SubsKind(p)
	if err != nil {
		return Subscription{}, err
	}

	return Subscription{
		ID:        orderID,
		AccountID: id,
		ListPrice: p.ListPrice,
		Amount:    p.NetPrice, // Modified for upgrade
		Coordinate: Coordinate{
			Tier:  p.Tier,
			Cycle: p.Cycle,
		},
		Currency:      p.Currency,
		CycleCount:    p.CycleCount, // Modified for upgrade
		ExtraDays:     p.ExtraDays,  // Modified for upgrade
		Usage:         kind,
		LastUpgradeID: null.String{}, // Only for upgrade
		PaymentMethod: method,
		//WxAppID:       null.String{}, // To be populated
		//StartDate:     chrono.Date{},
		//EndDate:       chrono.Date{},
		//CreatedAt:     chrono.Time{},
		//ConfirmedAt:   chrono.Time{},
	}, nil
}

// NewUpgradeOrder creates an upgrade order.
// Deprecate
func NewUpgradeOrder(accountID AccountID, up Upgrade) (Subscription, error) {
	id, err := GenerateOrderID()

	if err != nil {
		return Subscription{}, err
	}

	s := Subscription{
		ID:        id,
		AccountID: accountID,
		ListPrice: up.ListPrice,
		Amount:    up.NetPrice,
		Coordinate: Coordinate{
			Tier:  up.Tier,
			Cycle: up.Cycle,
		},
		Currency:   up.Currency,
		CycleCount: up.CycleCount,
		ExtraDays:  up.ExtraDays,
		Usage:      SubsKindUpgrade,
		CreatedAt:  chrono.TimeNow(),
	}

	return s, nil
}

func (s Subscription) WithUpgrade(source []BalanceSource) (Subscription, UpgradePreview) {
	up := NewUpgradePreview(source)
	up.OrderID = null.StringFrom(s.ID)

	s.Amount = up.Plan.NetPrice
	s.CycleCount = up.Plan.CycleCount
	s.ExtraDays = up.Plan.ExtraDays

	return s, up
}

// AliPrice converts Charged price to ailpay format
func (s Subscription) AliPrice() string {
	return strconv.FormatFloat(s.Amount, 'f', 2, 32)
}

// PriceInCent converts Charged price to int64 in cent for comparison with wx notification.
func (s Subscription) PriceInCent() int64 {
	return int64(s.Amount * 100)
}

func (s Subscription) ReadableAmount() string {
	return fmt.Sprintf("%s%.2f",
		strings.ToUpper(s.Currency),
		s.Amount,
	)
}

func (s Subscription) IsConfirmed() bool {
	return !s.ConfirmedAt.IsZero()
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
			return util2.ErrTierMismatched
		}
		if m.Tier == enum.InvalidTier {
			return util2.ErrNoUpgradingTarget
		}
		// For upgrading, order's tier must be different
		// from member's tier; otherwise this might be
		// a duplicate upgrading request.
		if m.Tier == s.Tier {
			return util2.ErrDuplicateUpgrading
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

	return s, nil
}

type ConfirmationResult struct {
	OrderID   string
	Succeeded bool
	Failed    null.String
	Retry     bool
}

func (r ConfirmationResult) Error() string {
	return r.Failed.String
}

func NewConfirmationSucceeded(orderID string) *ConfirmationResult {
	return &ConfirmationResult{
		OrderID:   orderID,
		Succeeded: true,
	}
}

func NewConfirmationFailed(orderID string, reason error, retry bool) *ConfirmationResult {
	return &ConfirmationResult{
		OrderID: orderID,
		Failed:  null.StringFrom(reason.Error()),
		Retry:   retry,
	}
}
