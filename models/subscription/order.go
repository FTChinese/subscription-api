package subscription

import (
	"fmt"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"strings"
	"time"

	"github.com/FTChinese/go-rest"
	"github.com/pkg/errors"
	"gitlab.com/ftchinese/subscription-api/models/reader"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

// GenerateOrderID creates an order memberID.
// The memberID has a total length of 18 chars.
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
	ID string `json:"memberID" db:"order_id"`
	reader.MemberID
	plan.BasePlan
	Price float64 `json:"price" db:"price"` // Price of a plan, prior to discount.
	Charge
	Duration
	Usage plan.SubsKind `json:"usageType" db:"usage_type"` // The usage of this order: creat new, renew, or upgrade?
	//LastUpgradeID null.String    `json:"-" db:"last_upgrade_id"`
	PaymentMethod   enum.PayMethod `json:"payMethod" db:"payment_method"`
	WxAppID         null.String    `json:"-" db:"wx_app_id"` // Wechat specific
	UpgradeSchemaID null.String    `json:"-" db:"upgrade_id"`
	CreatedAt       chrono.Time    `json:"createdAt" db:"created_at"`
	ConfirmedAt     chrono.Time    `json:"-" db:"confirmed_at"` // When the payment is confirmed.
	StartDate       chrono.Date    `json:"-" db:"start_date"`   // Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	EndDate         chrono.Date    `json:"-" db:"end_date"`     // Membership end date for this order. Depends on start date.

	MemberSnapshotID null.String `json:"-" db:"member_snapshot_id"` // Member data the moment this order is created. Null for a new member.
}

func (o Order) IsZero() bool {
	return o.ID == ""
}

// NewOrder creates a new subscription order.
// If later it is found that this order is used for upgrading,
// upgrade it and returns a new instance with upgrading price.
//func NewOrder(
//	id reader.MemberID,
//	p plan.Plan,
//	method enum.PayMethod,
//	kind plan.SubsKind,
//) (Order, error) {
//	orderID, err := GenerateOrderID()
//
//	if err != nil {
//		return Order{}, err
//	}
//
//	return Order{
//		ID:            orderID,
//		MemberID:      id,
//		Usage:         kind,
//		PaymentMethod: method,
//		CreatedAt:     chrono.TimeNow(),
//	}, nil
//}

//func (o Order) WithUpgrade(up UpgradeSchema) Order {
//
//	o.Amount = up.Plan.Amount
//	o.CycleCount = up.Plan.CycleCount
//	o.ExtraDays = up.Plan.ExtraDays
//	o.UpgradeSchemaID = null.StringFrom(up.ID)
//
//	return o
//}

// AliPrice converts Charged price to ailpay format
//func (o Order) AliPrice() string {
//	return strconv.FormatFloat(o.Amount, 'f', 2, 32)
//}

// AmountInCent converts Charged price to int64 in cent for comparison with wx notification.
// Deprecated:
func (o Order) AmountInCent() int64 {
	return int64(o.Amount * 100)
}

func (o Order) ReadableAmount() string {
	return fmt.Sprintf("%s%.2f",
		strings.ToUpper(o.Currency),
		o.Amount,
	)
}

func (o Order) IsConfirmed() bool {
	return !o.ConfirmedAt.IsZero()
}

func (o Order) getStartDate(m Membership, confirmedAt time.Time) time.Time {
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
		if o.Usage == plan.SubsKindUpgrade {
			startTime = confirmedAt
		} else {
			startTime = m.ExpireDate.Time
		}
	}

	return startTime
}

func (o Order) getEndDate(startTime time.Time) (time.Time, error) {
	var endTime time.Time

	switch o.Cycle {
	case enum.CycleYear:
		endTime = startTime.AddDate(int(o.CycleCount), 0, int(o.ExtraDays))

	case enum.CycleMonth:
		endTime = startTime.AddDate(0, int(o.CycleCount), int(o.ExtraDays))

	default:
		return endTime, errors.New("invalid billing cycle")
	}

	return endTime, nil
}

// Confirm updates an order with existing membership.
// Zero membership is a valid value.
//func (o Order) Confirm(m Membership, confirmedAt time.Time) (Order, error) {
//
//	startTime := o.getStartDate(m, confirmedAt)
//	endTime, err := o.getEndDate(startTime)
//	if err != nil {
//		return o, err
//	}
//
//	o.ConfirmedAt = chrono.TimeFrom(confirmedAt)
//	o.StartDate = chrono.DateFrom(startTime)
//	o.EndDate = chrono.DateFrom(endTime)
//
//	return o, nil
//}
