package ftcpay

import (
	"fmt"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/conv"
	"github.com/FTChinese/subscription-api/pkg/ids"
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
	StartDate   chrono.Date `db:"start_date"`
	EndDate     chrono.Date `db:"end_date"`
}

func (lo LockedOrder) IsConfirmed() bool {
	return !lo.ConfirmedAt.IsZero()
}

// Order contains the details of a user's action to place an order.
// This is the centrum of the whole subscription process.
// An order could represents 12 status of user:
// A user is allowed to to at max 2 ids - ftc or wechat, or both. This is 3 possible choices.
// A user could choose between 2 payment methods;
// An order could create, renew or upgrade a member.
// And tier + cycle have 3 combination.
// All those combinations add up to 3 * 2 * 3 * 3 = 54
type Order struct {
	// Fields common to all.
	ID string `json:"id" db:"order_id"`
	ids.UserIDs
	Tier          enum.Tier      `json:"tier" db:"tier"`
	Cycle         enum.Cycle     `json:"cycle" db:"cycle"` // Deprecated
	Kind          enum.OrderKind `json:"kind" db:"kind"`
	OriginalPrice float64        `json:"originalPrice" db:"original_price"` // The original price.
	PayableAmount float64        `json:"payableAmount" db:"payable_amount"`
	PaymentMethod enum.PayMethod `json:"payMethod" db:"payment_method"`
	YearsCount    int64          `json:"yearsCount" db:"years_count"`
	MonthsCount   int64          `json:"monthsCount" db:"months_count"`
	DaysCount     int64          `json:"daysCount" db:"days_count"`
	WxAppID       null.String    `json:"-" db:"wx_app_id"`
	ConfirmedAt   chrono.Time    `json:"confirmedAt" db:"confirmed_utc"` // When the payment is confirmed.
	CreatedAt     chrono.Time    `json:"createdAt" db:"created_utc"`
	// Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	StartDate chrono.Date `json:"startDate" db:"start_date"`
	// Membership end date for this order. Depends on start date.
	EndDate chrono.Date `json:"endDate" db:"end_date"`
}

func NewOrder(cart reader.ShoppingCart) (Order, error) {
	orderID, err := ids.OrderID()
	if err != nil {
		return Order{}, err
	}

	ymd := cart.FtcItem.PeriodCount()

	return Order{
		ID:            orderID,
		UserIDs:       cart.Account.CompoundIDs(),
		Tier:          cart.FtcItem.Price.Tier,
		Cycle:         ymd.EqCycle(),
		Kind:          cart.Intent.Kind.ToOrderKind(),
		OriginalPrice: cart.FtcItem.Price.UnitAmount,
		PayableAmount: cart.FtcItem.PayableAmount(),
		PaymentMethod: cart.PayMethod,
		YearsCount:    ymd.Years,
		MonthsCount:   ymd.Months,
		DaysCount:     ymd.Months,
		WxAppID:       cart.WxAppID,
		ConfirmedAt:   chrono.Time{},
		CreatedAt:     chrono.TimeNow(),
		StartDate:     chrono.Date{},
		EndDate:       chrono.Date{},
	}, nil
}

// MergeTail merges two orders.
// This is a hack due to previously MySQL refuse to retrieve
// the ftc_trade table exceeding a dozen columns.
func (o Order) MergeTail(t Order) Order {
	o.PaymentMethod = t.PaymentMethod
	o.YearsCount = t.YearsCount
	o.MonthsCount = t.MonthsCount
	o.DaysCount = t.DaysCount
	o.WxAppID = t.WxAppID
	o.CreatedAt = t.CreatedAt
	o.ConfirmedAt = t.ConfirmedAt
	o.StartDate = t.StartDate
	o.EndDate = t.EndDate

	return o
}

func (o Order) MergeLocked(l LockedOrder) Order {
	o.ConfirmedAt = l.ConfirmedAt
	o.StartDate = l.StartDate
	o.EndDate = l.EndDate

	return o
}

// PeriodCount produces a dt.YearMonthDay instance for easy calculation.
// An extra day is always given as a bonus.
func (o Order) PeriodCount() dt.YearMonthDay {
	return dt.YearMonthDay{
		Years:  o.YearsCount,
		Months: o.MonthsCount,
		Days:   o.DaysCount + 1,
	}
}

func (o Order) PaymentTitle() string {
	return price.BuildPaymentTitle(
		o.Kind,
		o.Tier,
		dt.YearMonthDay{
			Years:  o.YearsCount,
			Months: o.MonthsCount,
			Days:   o.DaysCount,
		})
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

func (o Order) AliPayable() string {
	return conv.FormatMoney(o.PayableAmount)
}

func (o Order) WxPayable() int64 {
	return conv.MoneyCent(o.PayableAmount)
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
	if conv.MoneyCent(o.PayableAmount) != result.Amount.Int64 {
		return fmt.Errorf("amount mismatched: expected: %d, actual: %d", o.WxPayable(), result.Amount.Int64)
	}

	return nil
}

// CalibratedKind changes order kind to renew in case
// it was created for upgrading while upon confirmation,
// membership already upgraded to premium.
// This situation is rare but possible under high concurrency.
func (o Order) CalibratedKind(m reader.Membership) enum.OrderKind {

	kind := calibrateOrderKind(m, price.Edition{
		Tier:  o.Tier,
		Cycle: o.Cycle,
	})
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

// Confirmed updates this order's ConfirmedAt field and set its starting and ending time.
// For addon order, the starting and ending time will be null since it is not used yet.
func (o Order) Confirmed(at chrono.Time, p dt.TimeSlot) Order {
	o.ConfirmedAt = at
	o.StartDate = chrono.DateFrom(p.StartUTC.Time)
	o.EndDate = chrono.DateFrom(p.EndUTC.Time)

	return o
}

type OrderList struct {
	Total int64 `json:"total" db:"row_count"`
	gorest.Pagination
	Data []Order `json:"data"`
	Err  error   `json:"-"`
}
