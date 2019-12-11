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
	live          bool // Determines the price.
	paymentResult PaymentResult
	membership    Membership // Current membership.
	order         Order      // The order corresponding to a webhook.
}

func NewConfirmationBuilder(result PaymentResult, live bool) *ConfirmationBuilder {
	return &ConfirmationBuilder{
		live:          live,
		paymentResult: result,
		membership:    Membership{},
		order:         Order{},
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

	if b.order.AmountInCent(b.live) != b.paymentResult.Amount {
		return &ConfirmError{
			Err:   fmt.Errorf("amount mismatched: expected: %d, actual: %d", b.order.AmountInCent(b.live), b.paymentResult.Amount),
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

func (b *ConfirmationBuilder) Build() (ConfirmationResult, error) {

	startTime := b.pickStartTime()
	if startTime.IsZero() {
		return ConfirmationResult{}, errors.New("cannot determine order start time")
	}

	endTime, err := b.order.getEndDate(startTime)
	if err != nil {
		return ConfirmationResult{}, err
	}

	snapshot := b.membership.Snapshot(b.order.Usage.SnapshotReason())

	order := b.confirmedOrder(startTime, endTime, snapshot.SnapshotID)

	m, err := b.membership.FromAliWxOrder(order)
	if err != nil {
		return ConfirmationResult{}, err
	}

	return ConfirmationResult{
		Order:      order,
		Membership: m,
		Snapshot:   snapshot,
	}, nil
}

func (b *ConfirmationBuilder) confirmedOrder(start time.Time, end time.Time, snapshotID string) Order {
	order := b.order

	order.ConfirmedAt = chrono.TimeFrom(b.paymentResult.ConfirmedAt)
	order.StartDate = chrono.DateFrom(start)
	order.EndDate = chrono.DateFrom(end)
	order.MemberSnapshotID = null.NewString(snapshotID, snapshotID != "")

	return order
}

func (b *ConfirmationBuilder) ConfirmedMembership(order Order) Membership {
	m := b.membership

	if b.membership.IsZero() {
		m.MemberID = order.MemberID
	}

	m.Tier = order.Tier
	m.Cycle = order.Cycle
	m.ExpireDate = chrono.DateFrom(order.EndDate.Time)
	m.PaymentMethod = order.PaymentMethod
	m.StripeSubID = null.String{}
	m.StripePlanID = null.String{}
	m.AutoRenewal = false

	return m
}
