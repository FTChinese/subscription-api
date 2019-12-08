package subscription

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"time"
)

// ConfirmationBuilder is used to confirm an
// order paid via alipay or wxpay.
// Those two providers does not support
// subscription mode, we need to turn their
// one-time purchase system to subscription.
// When webhook is triggered, a user's current
// membership needs to be inspected to see
// when shall the membership starts.
type ConfirmationBuilder struct {
	paymentResult PaymentResult
	membership    Membership // Current membership.
	order         Order      // The order corresponding to a webhook.

	startTime time.Time
	endTime   time.Time
}

func NewConfirmationBuilder(result PaymentResult) *ConfirmationBuilder {
	return &ConfirmationBuilder{
		paymentResult: result,
		membership:    Membership{},
		order:         Order{},
		startTime:     time.Time{},
		endTime:       time.Time{},
	}
}

func (b *ConfirmationBuilder) SetMembership(m Membership) *ConfirmationBuilder {
	b.membership = m
	return b
}

func (b *ConfirmationBuilder) SetOrder(o Order) *ConfirmationBuilder {
	b.order = o
	return b
}

func (b *ConfirmationBuilder) SetPaymentResult(r PaymentResult) *ConfirmationBuilder {
	b.paymentResult = r
	return b
}

// pickStartTime decides what is the start time
// for this purchase.
func (b *ConfirmationBuilder) pickStartTime() time.Time {
	// If current membership is already expired,
	// always use the order's confirmation time.
	// This includes empty membership.
	if b.membership.IsExpired() {
		return b.paymentResult.ConfirmedAt
	}

	// For upgrading, it always starts at
	// confirmation time
	if b.order.Usage == plan.SubsKindUpgrade {
		return b.paymentResult.ConfirmedAt
	}

	return b.membership.ExpireDate.Time
}

func (b *ConfirmationBuilder) ValidateOrder() *ConfirmError {

	if b.order.IsConfirmed() {
		return &ConfirmError{
			Err:   errors.New("subscription order is already confirmed"),
			Retry: false,
		}
	}

	if b.order.AmountInCent() != b.paymentResult.Amount {
		return &ConfirmError{
			Err:   fmt.Errorf("amount mismatched: expected: %d, actual: %d", b.order.AmountInCent(), b.paymentResult.Amount),
			Retry: false,
		}
	}

	return nil
}

func (b *ConfirmationBuilder) ValidateDuplicateUpgrading() *ConfirmError {
	if b.order.Usage == plan.SubsKindUpgrade && b.membership.IsValidPremium() {
		return &ConfirmError{
			Err:   errors.New("duplicate upgrading"),
			Retry: false,
		}
	}

	return nil
}

func (b *ConfirmationBuilder) Build() error {

	b.startTime = b.pickStartTime()
	end, err := b.order.getEndDate(b.startTime)
	if err != nil {
		return err
	}

	b.endTime = end

	return nil
}

func (b *ConfirmationBuilder) ConfirmedOrder() Order {
	order := b.order

	order.ConfirmedAt = chrono.TimeFrom(b.paymentResult.ConfirmedAt)
	order.StartDate = chrono.DateFrom(b.startTime)
	order.EndDate = chrono.DateFrom(b.endTime)

	return order
}

func (b *ConfirmationBuilder) ConfirmedMembership() Membership {
	m := b.membership

	m.GenerateID()

	if b.membership.IsZero() {
		m.MemberID = b.order.MemberID
	}

	m.Tier = b.order.Tier
	m.Cycle = b.order.Cycle
	m.ExpireDate = b.order.EndDate
	m.PaymentMethod = b.order.PaymentMethod
	m.StripeSubID = null.String{}
	m.StripePlanID = null.String{}
	m.AutoRenewal = false

	return m
}
