package subrepo

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"time"
)

// FreeUpgrade upgrade a membership to premium for free if user's current balance
// is large enough to cover premium's price.
// This won't happen based on current restrictions of max renewal for 3 consecutive years.
// It is provided here just for logical completeness.
func (env Env) FreeUpgrade(builder *subs.OrderBuilder) (subs.ConfirmationResult, error) {

	tx, err := env.BeginOrderTx()
	if err != nil {
		return subs.ConfirmationResult{}, err
	}

	member, err := tx.RetrieveMember(builder.GetReaderID())
	if err != nil {
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, err
	}

	err = builder.DeduceSubsKind(member)
	if err != nil {
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, err
	}

	if builder.GetSubsKind() != enum.OrderKindUpgrade {
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, subs.ErrUpgradeInvalid
	}

	orders, err := tx.FindBalanceSources(builder.GetReaderID())
	if err != nil {
		return subs.ConfirmationResult{}, err
	}

	wallet := subs.NewWallet(orders, time.Now())

	builder.SetWallet(wallet)

	if err := builder.Build(); err != nil {
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, err
	}

	order, _ := builder.GetOrder()

	if !order.IsFreeUpgrade() {
		return subs.ConfirmationResult{}, errors.New("current balance is not enough to cover upgrading cost")
	}

	confirmed, err := subs.NewConfirmationBuilder(subs.PaymentResult{
		Amount:      0,
		OrderID:     order.ID,
		ConfirmedAt: chrono.TimeNow(),
	}).
		SetMembership(member).
		SetOrder(order).
		Build()

	if err != nil {
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, err
	}

	// Save order
	if err := tx.SaveOrder(confirmed.Order); err != nil {
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, err
	}

	// Save updated membership.
	if err := tx.UpdateMember(confirmed.Membership); err != nil {
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, err
	}

	if err := tx.SaveProratedOrders(builder.GetWallet().Sources); err != nil {
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return subs.ConfirmationResult{}, err
	}

	return confirmed, nil
}
