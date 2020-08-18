package builder

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
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
	paymentResult subs.PaymentResult
	membership    subs.Membership // Current membership.
	order         subs.Order      // The order corresponding to a webhook.
}

func NewConfirmationBuilder(result subs.PaymentResult, live bool) *ConfirmationBuilder {
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

	order, err := b.order.Confirm(b.membership, b.paymentResult.ConfirmedAt)
	if err != nil {
		return subs.ConfirmationResult{}, err
	}

	return subs.ConfirmationResult{
		Order:      order,
		Membership: order.Membership(),
		Snapshot: subs.MemberSnapshot{
			SnapshotID: subs.GenerateSnapshotID(),
			Reason:     subs.GetSnapshotReason(order.Usage),
			CreatedUTC: chrono.TimeNow(),
			OrderID:    null.StringFrom(order.ID),
			Membership: b.membership,
		},
	}, nil
}
