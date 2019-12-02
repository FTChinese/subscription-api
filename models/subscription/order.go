package subscription

import (
	"fmt"
	"gitlab.com/ftchinese/subscription-api/models/plan"
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
	reader.MemberID
	plan.Plan
	Usage SubsKind `json:"usageType" db:"usage_type"` // The usage of this order: creat new, renew, or upgrade?
	//LastUpgradeID null.String    `json:"-" db:"last_upgrade_id"`
	PaymentMethod    enum.PayMethod `json:"payMethod" db:"payment_method"`
	WxAppID          null.String    `json:"-" db:"wx_app_id"`  // Wechat specific
	StartDate        chrono.Date    `json:"-" db:"start_date"` // Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	EndDate          chrono.Date    `json:"-" db:"end_date"`   // Membership end date for this order. Depends on start date.
	CreatedAt        chrono.Time    `json:"createdAt" db:"created_at"`
	ConfirmedAt      chrono.Time    `json:"-" db:"confirmed_at"` // When the payment is confirmed.
	UpgradeID        null.String    `json:"-" db:"upgrade_id"`
	MemberSnapshotID null.String    `json:"-" db:"member_snapshot_id"` // Member data the moment this order is created. Null for a new member.
}

// NewOrder creates a new subscription order.
// If later it is found that this order is used for upgrading,
// upgrade it and returns a new instance with upgrading price.
func NewOrder(
	id reader.MemberID,
	p plan.Plan,
	method enum.PayMethod,
	kind SubsKind,
) (Order, error) {
	orderID, err := GenerateOrderID()

	if err != nil {
		return Order{}, err
	}

	return Order{
		ID:            orderID,
		MemberID:      id,
		Plan:          p,
		Usage:         kind,
		PaymentMethod: method,
		CreatedAt:     chrono.TimeNow(),
	}, nil
}

func NewFreeUpgradeOrder(id reader.MemberID, up plan.UpgradeIntent) (Order, error) {

	startTime := time.Now()
	endTime, err := up.Plan.GetPeriodEnd(startTime)
	if err != nil {
		return Order{}, err
	}

	order, err := NewOrder(
		id,
		up.Plan,
		enum.PayMethodNull,
		SubsKindUpgrade)

	if err != nil {
		return order, err
	}

	order.StartDate = chrono.DateFrom(startTime)
	order.EndDate = chrono.DateFrom(endTime)
	order.CreatedAt = chrono.TimeNow()
	order.ConfirmedAt = chrono.TimeNow()
	order.UpgradeID = null.StringFrom(up.ID)

	return order, nil
}

func (s Order) WithUpgrade(up plan.UpgradeIntent) Order {

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
		strings.ToUpper(s.Currency),
		s.Amount,
	)
}

func (s Order) IsConfirmed() bool {
	return !s.ConfirmedAt.IsZero()
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
