package builder

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/models/subscription"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/guregu/null"
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
	paymentResult subscription.PaymentResult
	membership    subs.Membership // Current membership.
	order         subs.Order      // The order corresponding to a webhook.
}

func NewConfirmationBuilder(result subscription.PaymentResult, live bool) *ConfirmationBuilder {
	return &ConfirmationBuilder{
		live:          live,
		paymentResult: result,
		membership:    subs.Membership{},
		order:         subs.Order{},
	}
}

func (b *ConfirmationBuilder) SetMembership(m subs.Membership) *ConfirmationBuilder {
	b.membership = m
	return b
}

func (b *ConfirmationBuilder) SetOrder(o subs.Order) *ConfirmationBuilder {
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
	if b.order.Usage == enum.OrderKindUpgrade {
		return b.paymentResult.ConfirmedAt
	}

	return b.membership.ExpireDate.Time
}

func (b *ConfirmationBuilder) ValidateOrder() *subs.ConfirmError {

	if b.order.IsConfirmed() {
		return &subs.ConfirmError{
			Err:   errors.New("subscription order is already confirmed"),
			Retry: false,
		}
	}

	if b.order.AmountInCent(b.live) != b.paymentResult.Amount {
		return &subs.ConfirmError{
			Err:   fmt.Errorf("amount mismatched: expected: %d, actual: %d", b.order.AmountInCent(b.live), b.paymentResult.Amount),
			Retry: false,
		}
	}

	return nil
}

func (b *ConfirmationBuilder) ValidateDuplicateUpgrading() *subs.ConfirmError {
	if b.order.Usage == enum.OrderKindUpgrade && b.membership.IsValidPremium() {
		return &subs.ConfirmError{
			Err:   errors.New("duplicate upgrading"),
			Retry: false,
		}
	}

	return nil
}

func (b *ConfirmationBuilder) Build() (subs.ConfirmationResult, error) {

	startTime := b.pickStartTime()
	if startTime.IsZero() {
		return subs.ConfirmationResult{}, errors.New("cannot determine order start time")
	}

	endTime, err := b.order.getEndDate(startTime)
	if err != nil {
		return subs.ConfirmationResult{}, err
	}

	snapshot := b.membership.Snapshot(subs.GetSnapshotReason(b.order.Usage))

	order := b.confirmedOrder(startTime, endTime, snapshot.SnapshotID)

	m, err := b.membership.FromAliWxOrder(order)
	if err != nil {
		return subs.ConfirmationResult{}, err
	}

	return subs.ConfirmationResult{
		Order:      order,
		Membership: m,
		Snapshot:   snapshot,
	}, nil
}

func (b *ConfirmationBuilder) confirmedOrder(start time.Time, end time.Time, snapshotID string) subs.Order {
	order := b.order

	order.ConfirmedAt = chrono.TimeFrom(b.paymentResult.ConfirmedAt)
	order.StartDate = chrono.DateFrom(start)
	order.EndDate = chrono.DateFrom(end)
	order.MemberSnapshotID = null.NewString(snapshotID, snapshotID != "")

	return order
}

func (b *ConfirmationBuilder) ConfirmedMembership(order subs.Order) subs.Membership {
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
	m.AutoRenew = false

	return m
}
