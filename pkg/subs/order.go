package subs

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
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
	PlanID     string      `json:"planId" db:"plan_id"`
	DiscountID null.String `json:"discountId" db:"discount_id"`
	Price      float64     `json:"price" db:"price"` // Price of a plan, prior to discount.
	product.Edition
	product.Charge
	product.Duration
	Kind enum.OrderKind `json:"usageType" db:"kind"` // The usage of this order: creat new, renew, or upgrade?
	//LastUpgradeID null.String    `json:"-" db:"last_upgrade_id"`
	PaymentMethod enum.PayMethod `json:"payMethod" db:"payment_method"`
	TotalBalance  null.Float     `json:"totalBalance" db:"total_balance"` // Only for upgrade
	WxAppID       null.String    `json:"-" db:"wx_app_id"`                // Wechat specific. Used by webhook to verify notification.
	CreatedAt     chrono.Time    `json:"createdAt" db:"created_utc"`
	ConfirmedAt   chrono.Time    `json:"confirmedAt" db:"confirmed_utc"` // When the payment is confirmed.
	PurchasedPeriod
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

func (o Order) ValidatePayment(result PaymentResult) error {
	if o.AmountInCent() != result.Amount.Int64 {
		return fmt.Errorf("amount mismatched: expected: %d, actual: %d", o.AmountInCent(), result.Amount.Int64)
	}

	return nil
}

func (o Order) ValidateDupUpgrade(m reader.Membership) error {
	if o.Kind == enum.OrderKindUpgrade && m.IsValidPremium() {
		return errors.New("duplicate upgrading")
	}

	return nil
}

// pick which date to use as start date upon confirmation.
// expireDate refers to current membership's expireDate.
func (o Order) pickStartDate(expireDate chrono.Date) chrono.Date {
	// If this is an upgrade order, or membership is expired, use confirmation time.
	if o.Kind == enum.OrderKindUpgrade || o.ConfirmedAt.Time.After(expireDate.Time) {
		return chrono.DateFrom(o.ConfirmedAt.Time)
	}

	return expireDate
}

// Membership build a membership based on this order.
// The order must be already confirmed.
func (o Order) Membership() (reader.Membership, error) {
	if !o.IsConfirmed() {
		return reader.Membership{}, fmt.Errorf("order %s used to build membership is not confirmed yet", o.ID)
	}

	return reader.Membership{
		MemberID:      o.MemberID,
		Edition:       o.Edition,
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    o.EndDate,
		PaymentMethod: o.PaymentMethod,
		FtcPlanID:     null.StringFrom(o.PlanID),
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        enum.SubsStatusNull,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
	}, nil
}

// Confirm an order based on existing membership.
// If current membership is not expired, the order's
// purchased start date starts from the membership's
// expiration date; otherwise it starts from the
// confirmation time received by webhook.
// If this order is used for upgrading, it always starts
// at now.
func (o Order) Confirm(pr PaymentResult, m reader.Membership) (ConfirmationResult, error) {
	o.ConfirmedAt = chrono.TimeFrom(pr.ConfirmedAt.Time)

	period, err := NewPeriodBuilder(
		o.Edition,
		o.Duration).
		Build(o.pickStartDate(m.ExpireDate))
	if err != nil {
		return ConfirmationResult{}, err
	}

	o.PurchasedPeriod = period

	newMember, err := o.Membership()
	if err != nil {
		return ConfirmationResult{}, err
	}

	snapshot := m.Snapshot(reader.FtcArchiver(o.Kind))
	if !m.IsZero() {
		snapshot = snapshot.WithOrder(o.ID)
	}

	return ConfirmationResult{
		Order:      o,
		Membership: newMember,
		Snapshot:   snapshot,
	}, nil
}
