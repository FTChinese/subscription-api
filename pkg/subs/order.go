package subs

import (
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

// LockedOrder lock a row of order and retrieves the minimal data.
// This is used to resolve an unknown server problem that
// when the retrieved data exceed a certain amount, MySQL
// does not respond.
type LockedOrder struct {
	ID          string      `db:"order_id"`
	ConfirmedAt chrono.Time `db:"confirmed_utc"`
	dt.DatePeriod
}

func (lo LockedOrder) IsConfirmed() bool {
	return !lo.ConfirmedAt.IsZero()
}

// Merge updates an order retrieved outside a transaction in case
// the full order is not confirmed but the locked version is changed.
// This is used to solved concurrency issue.
func (lo LockedOrder) Merge(o Order) Order {
	o.ConfirmedAt = lo.ConfirmedAt
	o.DatePeriod = lo.DatePeriod

	return o
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
	pkg.UserIDs
	PlanID     string      `json:"priceId" db:"plan_id"`
	DiscountID null.String `json:"discountId" db:"discount_id"`
	Price      float64     `json:"price" db:"price"` // Price of a plan, prior to discount.
	price.Edition
	price.Charge
	Kind          enum.OrderKind `json:"kind" db:"kind"` // The usage of this order: creat new, renew, or upgrade?
	PaymentMethod enum.PayMethod `json:"payMethod" db:"payment_method"`
	WxAppID       null.String    `json:"-" db:"wx_app_id"` // Wechat specific. Used by webhook to verify notification.
	CreatedAt     chrono.Time    `json:"createdAt" db:"created_utc"`
	ConfirmedAt   chrono.Time    `json:"confirmedAt" db:"confirmed_utc"` // When the payment is confirmed.
	dt.DatePeriod
	LiveMode bool `json:"liveMode"`
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

// IsExpireDateSynced tests whether a confirmed order and the end date is
// transferred to membership.
func (o Order) IsExpireDateSynced(m reader.Membership) bool {
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

	// Order kinds decides what fields to compare:
	// For Create, Renew, Upgrade, we change membership's expiration date;
	// For add-ons, we change the AddOn.
	// Every Upgrade and AddOn order will generate a row in the addon table.
	// We have no way to know if add-on is created unless we query
	// db.
	if o.Kind == enum.OrderKindAddOn {
		return true
	}

	// As long as membership's expiration date is equal or after
	// the order's end time, we think the order is synced.
	return !m.ExpireDate.Before(o.EndDate.Time)
}

func (o Order) ValidatePayment(result PaymentResult) error {
	if o.AmountInCent() != result.Amount.Int64 {
		return fmt.Errorf("amount mismatched: expected: %d, actual: %d", o.AmountInCent(), result.Amount.Int64)
	}

	return nil
}

// CalibratedKind changes order kind to renew in case
// it was created for upgrading while upon confirmation,
// membership already upgraded to premium.
// This situation is rare but possible under high concurrency.
func (o Order) CalibratedKind(m reader.Membership) enum.OrderKind {

	kind := calibrateOrderKind(m, o.Edition)
	if kind == enum.OrderKindNull {
		return o.Kind
	}

	return kind
}

// Adjust order kind upon confirmation since it might be different from
// the one when creating an order due to concurrency.
func calibrateOrderKind(m reader.Membership, e price.Edition) enum.OrderKind {
	if m.IsExpired() {
		return enum.OrderKindCreate
	}

	if m.IsInvalidStripe() {
		return enum.OrderKindCreate
	}

	switch m.PaymentMethod {
	case enum.PayMethodAli, enum.PayMethodWx:
		// For same tier, it is renewal.
		if m.Tier == e.Tier {
			return enum.OrderKindRenew
		}

		// Purchasing a different tier.
		switch e.Tier {
		// Standard to premium
		case enum.TierPremium:
			return enum.OrderKindUpgrade
		// Premium to standard.
		case enum.TierStandard:
			return enum.OrderKindAddOn
		}

	// If current membership comes Stripe or IAP,
	// it doesn't matter whatever user purchased
	// since you have to accept it.
	case enum.PayMethodStripe, enum.PayMethodApple, enum.PayMethodB2B:
		return enum.OrderKindAddOn
	}

	return enum.OrderKindNull
}

func (o Order) Confirmed(at chrono.Time, period dt.DateTimePeriod) Order {
	o.ConfirmedAt = at
	o.DatePeriod = period.ToDatePeriod()

	return o
}
