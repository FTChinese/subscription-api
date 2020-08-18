package subrepo

import (
	"github.com/FTChinese/subscription-api/models/subscription"
	"github.com/FTChinese/subscription-api/pkg/builder"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"time"
)

// FreeUpgrade upgrade a membership to premium for free if user's current balance
// is large enough to cover premium's price.
// This won't happen based on current restrictions of max renewal for 3 consecutive years.
// It is provided here just for logical completeness.
func (env SubEnv) FreeUpgrade(builder *builder.OrderBuilder) (subs.ConfirmationResult, error) {

	tx, err := env.BeginOrderTx()
	if err != nil {
		return subs.ConfirmationResult{}, err
	}

	member, err := tx.RetrieveMember(builder.GetReaderID())
	if err != nil {
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, err
	}

	if !member.PermitAliWxUpgrade() {
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, subscription.ErrUpgradeInvalid
	}

	orders, err := tx.FindBalanceSources(builder.GetReaderID())
	if err != nil {
		return subs.ConfirmationResult{}, err
	}

	wallet := subs.NewWallet(orders, time.Now())

	builder.SetMembership(member).
		SetWallet(wallet)

	if err := builder.Build(); err != nil {
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, err
	}

	confirmed, err := builder.FreeUpgrade()
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

	// Save upgrading schema.
	upgrade, _ := builder.UpgradeSchema()
	if err := tx.SaveUpgradeSchema(upgrade); err != nil {
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return subs.ConfirmationResult{}, err
	}

	return confirmed, nil
}
