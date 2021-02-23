package subs

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

const trialDays = 1

type LockedOrder struct {
	ID          string      `db:"order_id"`
	ConfirmedAt chrono.Time `db:"confirmed_utc"`
}

func (o LockedOrder) IsConfirmed() bool {
	return !o.ConfirmedAt.IsZero()
}

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
	PlanID     string      `json:"planId" db:"plan_id"`
	DiscountID null.String `json:"discountId" db:"discount_id"`
	Price      float64     `json:"price" db:"price"` // Price of a plan, prior to discount.
	price.Edition
	price.Charge
	Kind          enum.OrderKind `json:"kind" db:"kind"` // The usage of this order: creat new, renew, or upgrade?
	PaymentMethod enum.PayMethod `json:"payMethod" db:"payment_method"`
	WxAppID       null.String    `json:"-" db:"wx_app_id"` // Wechat specific. Used by webhook to verify notification.
	CreatedAt     chrono.Time    `json:"createdAt" db:"created_utc"`
	ConfirmedAt   chrono.Time    `json:"confirmedAt" db:"confirmed_utc"` // When the payment is confirmed.
	dt.DateRange
	LiveMode bool `json:"live"`
}

func (o Order) IsZero() bool {
	return o.ID == ""
}

func (o Order) IsConfirmed() bool {
	return !o.ConfirmedAt.IsZero()
}

func (o Order) IsAliWxPay() bool {
	return o.PaymentMethod == enum.PayMethodAli || o.PaymentMethod == enum.PayMethodWx
}

// IsSynced tests whether an order is confirmed and the end date is
// transferred to membership.
func (o Order) IsSynced(m reader.Membership) bool {
	if m.IsZero() {
		return false
	}

	if o.ConfirmedAt.IsZero() {
		return false
	}

	// As long the current membership expiration date is not before
	// this order's end date, we think the order is already synced to membership.
	if !o.IsAliWxPay() {
		return true
	}

	if o.EndDate.IsZero() {
		return false
	}

	return !m.ExpireDate.Before(o.EndDate.Time)
}

func (o Order) ValidatePayment(result PaymentResult) error {
	if o.AmountInCent() != result.Amount.Int64 {
		return fmt.Errorf("amount mismatched: expected: %d, actual: %d", o.AmountInCent(), result.Amount.Int64)
	}

	return nil
}

func (o Order) ToAddOn() addon.AddOn {
	return addon.AddOn{
		ID:            db.AddOnID(),
		Edition:       o.Edition,
		CycleCount:    1,
		DaysRemained:  trialDays,
		PaymentMethod: o.PaymentMethod,
		CompoundID:    o.CompoundID,
		OrderID:       null.StringFrom(o.ID),
		PlanID:        null.StringFrom(o.PlanID),
		CreatedUTC:    chrono.TimeNow(),
		ConsumedUTC:   chrono.Time{},
	}
}
