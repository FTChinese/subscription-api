package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"time"
)

// FreeUpgrade upgrade a membership to premium for free if user's current balance
// is large enough to cover premium's price.
// This won't happen based on current restrictions of max renewal for 3 consecutive years.
// It is provided here just for logical completeness.
func (env SubEnv) FreeUpgrade(builder *subscription.OrderBuilder) (subscription.ConfirmationResult, error) {

	tx, err := env.BeginOrderTx()
	if err != nil {
		return subscription.ConfirmationResult{}, err
	}

	member, err := tx.RetrieveMember(builder.GetReaderID())
	if err != nil {
		_ = tx.Rollback()
		return subscription.ConfirmationResult{}, err
	}

	if !member.PermitAliWxUpgrade() {
		_ = tx.Rollback()
		return subscription.ConfirmationResult{}, subscription.ErrUpgradeInvalid
	}

	orders, err := tx.FindBalanceSources(builder.GetReaderID())
	if err != nil {
		return subscription.ConfirmationResult{}, err
	}

	wallet := subscription.NewWallet(orders, time.Now())

	builder.SetMembership(member).
		SetWallet(wallet)

	if err := builder.Build(); err != nil {
		_ = tx.Rollback()
		return subscription.ConfirmationResult{}, err
	}

	confirmed, err := builder.FreeUpgrade()
	if err != nil {
		_ = tx.Rollback()
		return subscription.ConfirmationResult{}, err
	}

	// Save order
	if err := tx.SaveOrder(confirmed.Order); err != nil {
		_ = tx.Rollback()
		return subscription.ConfirmationResult{}, err
	}

	// Save updated membership.
	if err := tx.UpdateMember(confirmed.Membership); err != nil {
		_ = tx.Rollback()
		return subscription.ConfirmationResult{}, err
	}

	// Save upgrading schema.
	upgrade, _ := builder.UpgradeSchema()
	if err := tx.SaveUpgradeBalance(upgrade.UpgradeBalanceSchema); err != nil {
		_ = tx.Rollback()
		return subscription.ConfirmationResult{}, err
	}

	if err := tx.SaveProratedOrders(upgrade.Sources); err != nil {
		_ = tx.Rollback()
		return subscription.ConfirmationResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return subscription.ConfirmationResult{}, err
	}

	return confirmed, nil
}
