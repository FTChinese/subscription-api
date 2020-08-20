package subs

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

// FreeUpgrade contains the data after creating/upgrading/renewing a membership.
// All data here is in a final state.
// Those data can be directly save into database.
// The generated of those fields has a intertwined dependency
// on each other, so they are return in one batch.
type ConfirmationResult struct {
	Order      Order          // The confirmed order.
	Membership Membership     // The updated membership.
	Snapshot   MemberSnapshot // // Snapshot of previous membership
}

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
	if b.order.Kind == enum.OrderKindUpgrade && b.membership.IsValidPremium() {
		return &ConfirmError{
			Err:   errors.New("duplicate upgrading"),
			Retry: false,
		}
	}

	return nil
}

func (b *ConfirmationBuilder) Build() (ConfirmationResult, error) {

	order, err := b.order.Confirm(b.membership, b.paymentResult.ConfirmedAt)
	if err != nil {
		return ConfirmationResult{}, err
	}

	return ConfirmationResult{
		Order:      order,
		Membership: order.Membership(),
		Snapshot: MemberSnapshot{
			SnapshotID: GenerateSnapshotID(),
			Reason:     GetSnapshotReason(order.Kind),
			CreatedUTC: chrono.TimeNow(),
			OrderID:    null.StringFrom(order.ID),
			Membership: b.membership,
		},
	}, nil
}
