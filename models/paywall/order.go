package paywall

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	gorest "github.com/FTChinese/go-rest"
	"github.com/pkg/errors"
	"gitlab.com/ftchinese/subscription-api/models/reader"

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

// Subscription contains the details of a user's action to place an order.
// This is the centrum of the whole subscription process.
// An order could represents 12 status of user:
// A user is allowed to to at max 2 ids - ftc or wechat, or both. This is 3 possible choices.;
// A user could choose between 2 payment methods;
// An order could create, renew or upgrade a member.
// And tier + cycle have 3 combination.
// All those combination add up to 3 * 2 * 3 * 3 = 54
type Order struct {
	// Fields common to all.
	ID string `json:"id" db:"order_id"`
	reader.AccountID
	//Charge
	ListPrice float64 `json:"price" db:"price"`   // Price of a plan, prior to discount.
	Amount    float64 `json:"amount" db:"amount"` // Actually paid amount.
	Coordinate
	Currency   null.String `json:"-"`
	CycleCount int64       `json:"cycleCount" db:"cycle_count"` // Default to 1. Change it for upgrade
	ExtraDays  int64       `json:"extraDays" db:"extra_days"`   // Default to 1. Change it for upgraded.
	Usage      SubsKind    `json:"usageType" db:"usage_type"`   // The usage of this order: creat new, renew, or upgrade?
	//LastUpgradeID null.String    `json:"-" db:"last_upgrade_id"`
	PaymentMethod enum.PayMethod `json:"payMethod" db:"payment_method"`
	WxAppID       null.String    `json:"-" db:"wx_app_id"`  // Wechat specific
	StartDate     chrono.Date    `json:"-" db:"start_date"` // Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	EndDate       chrono.Date    `json:"-" db:"end_date"`   // Membership end date for this order. Depends on start date.
	//User          AccountID      `json:"-"`                 // Deprecate
	CreatedAt        chrono.Time `json:"createdAt" db:"created_at"`
	ConfirmedAt      chrono.Time `json:"-" db:"confirmed_at"` // When the payment is confirmed.
	UpgradeID        null.String `json:"-" db:"upgrade_id"`
	MemberSnapshotID null.String `json:"-" db:"member_snapshot_id"` // Member data the moment this order is created. Null for a new member.
}

// NewOrder creates a new subscription order.
// If later it is found that this order is used for upgrading,
// upgrade it and returns a new instance with upgrading price.
func NewOrder(
	id reader.AccountID,
	p Plan,
	method enum.PayMethod,
	m Membership,
) (Order, error) {
	orderID, err := GenerateOrderID()

	if err != nil {
		return Order{}, err
	}

	kind, err := m.SubsKind(p)
	if err != nil {
		return Order{}, err
	}

	return Order{
		ID:        orderID,
		AccountID: id,
		ListPrice: p.ListPrice,
		Amount:    p.NetPrice, // Modified for upgrade
		Coordinate: Coordinate{
			Tier:  p.Tier,
			Cycle: p.Cycle,
		},
		Currency:      null.StringFrom(p.Currency),
		CycleCount:    p.CycleCount, // Modified for upgrade
		ExtraDays:     p.ExtraDays,  // Modified for upgrade
		Usage:         kind,
		PaymentMethod: method,
		//WxAppID:       null.String{}, // To be populated
		//StartDate:     chrono.Date{},
		//EndDate:       chrono.Date{},
		CreatedAt: chrono.TimeNow(),
		//ConfirmedAt:   chrono.Time{},
		UpgradeID:        null.String{},
		MemberSnapshotID: null.String{},
	}, nil
}

func NewFreeUpgradeOrder(id reader.AccountID, up UpgradePlan) (Order, error) {
	orderID, err := GenerateOrderID()
	if err != nil {
		return Order{}, err
	}

	startTime := time.Now()
	endTime, err := up.Plan.GetPeriodEnd(startTime)
	if err != nil {
		return Order{}, err
	}

	return Order{
		ID:        orderID,
		AccountID: id,
		ListPrice: 0,
		Amount:    0,
		Coordinate: Coordinate{
			Tier:  up.Plan.Tier,
			Cycle: up.Plan.Cycle,
		},
		Currency:         null.StringFrom(up.Plan.Currency),
		CycleCount:       up.Plan.CycleCount,
		ExtraDays:        up.Plan.ExtraDays,
		Usage:            SubsKindUpgrade,
		PaymentMethod:    enum.InvalidPay,
		WxAppID:          null.String{},
		StartDate:        chrono.DateFrom(startTime),
		EndDate:          chrono.DateFrom(endTime),
		CreatedAt:        chrono.TimeNow(),
		ConfirmedAt:      chrono.TimeNow(),
		UpgradeID:        null.StringFrom(up.ID),
		MemberSnapshotID: null.String{}, // Manually add it later.
	}, nil
}

func (s Order) WithUpgrade(up UpgradePlan) Order {

	s.Amount = up.Plan.NetPrice
	s.CycleCount = up.Plan.CycleCount
	s.ExtraDays = up.Plan.ExtraDays
	s.UpgradeID = null.StringFrom(up.ID)

	return s
}

// AliPrice converts Charged price to ailpay format
func (s Order) AliPrice() string {
	return strconv.FormatFloat(s.Amount, 'f', 2, 32)
}

// AmountInCent converts Charged price to int64 in cent for comparison with wx notification.
func (s Order) AmountInCent() int64 {
	return int64(s.Amount * 100)
}

func (s Order) ReadableAmount() string {
	return fmt.Sprintf("%s%.2f",
		strings.ToUpper(s.Currency.String),
		s.Amount,
	)
}

func (s Order) IsConfirmed() bool {
	return !s.ConfirmedAt.IsZero()
}

func (s Order) GetAccountID() reader.AccountID {
	return reader.AccountID{
		CompoundID: s.CompoundID,
		FtcID:      s.FtcID,
		UnionID:    s.UnionID,
	}
}

func (s Order) getStartDate(m Membership, confirmedAt time.Time) time.Time {
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

func (s Order) getEndDate(startTime time.Time) (time.Time, error) {
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

// Confirm updates an order with existing membership.
// Zero membership is a valid value.
func (s Order) Confirm(m Membership, confirmedAt time.Time) (Order, error) {

	startTime := s.getStartDate(m, confirmedAt)
	endTime, err := s.getEndDate(startTime)
	if err != nil {
		return s, err
	}

	s.ConfirmedAt = chrono.TimeFrom(confirmedAt)
	s.StartDate = chrono.DateFrom(startTime)
	s.EndDate = chrono.DateFrom(endTime)

	return s, nil
}

// ConfirmationResult logs the result of confirmation.
type ConfirmationResult struct {
	OrderID   string
	Succeeded bool
	Failed    null.String
	Retry     bool
}

func (r ConfirmationResult) Error() string {
	return r.Failed.String
}

// NewConfirmationSucceeded createa a new instance of ConfirmationResult for success.
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
