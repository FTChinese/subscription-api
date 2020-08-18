package subs

import (
	"github.com/FTChinese/subscription-api/pkg/product"
	"time"

	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/pkg/errors"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

// Subscription contains the details of a user's action to place an order.
// This is the centrum of the whole subscription process.
// An order could represents 12 status of user:
// A user is allowed to to at max 2 ids - ftc or wechat, or both. This is 3 possible choices.
// A user could choose between 2 payment methods;
// An order could create, renew or upgrade a member.
// And tier + cycle have 3 combination.
// All those combination add up to 3 * 2 * 3 * 3 = 54
type Order struct {
	// Fields common to all.
	ID string `json:"id" db:"order_id"`
	reader.MemberID
	PlanID string `json:"plan_id" db:"plan_id"`
	product.Edition
	Price float64 `json:"price" db:"price"` // Price of a plan, prior to discount.
	Charge
	Duration
	Usage enum.OrderKind `json:"usageType" db:"kind"` // The usage of this order: creat new, renew, or upgrade?
	//LastUpgradeID null.String    `json:"-" db:"last_upgrade_id"`
	PaymentMethod enum.PayMethod `json:"payMethod" db:"payment_method"`
	WxAppID       null.String    `json:"-" db:"wx_app_id"` // Wechat specific
	CreatedAt     chrono.Time    `json:"createdAt" db:"created_at"`
	ConfirmedAt   chrono.Time    `json:"-" db:"confirmed_at"` // When the payment is confirmed.
	StartDate     chrono.Date    `json:"-" db:"start_date"`   // Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	EndDate       chrono.Date    `json:"-" db:"end_date"`     // Membership end date for this order. Depends on start date.
}

func (o Order) IsZero() bool {
	return o.ID == ""
}

func (o Order) IsConfirmed() bool {
	return !o.ConfirmedAt.IsZero()
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

func (o Order) SnapshotReason() enum.SnapshotReason {
	switch o.Usage {
	case enum.OrderKindRenew:
		return enum.SnapshotReasonRenew

	case enum.OrderKindUpgrade:
		return enum.SnapshotReasonUpgrade

	default:
		return enum.SnapshotReasonNull
	}
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
