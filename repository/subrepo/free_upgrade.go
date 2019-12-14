package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"time"
)

func (env SubEnv) FreeUpgrade(builder *subscription.OrderBuilder) (subscription.Order, error) {

	tx, err := env.BeginOrderTx()
	if err != nil {
		return subscription.Order{}, err
	}

	member, err := tx.RetrieveMember(builder.GetReaderID())
	if err != nil {
		_ = tx.Rollback()
		return subscription.Order{}, err
	}

	if !member.PermitAliWxUpgrade() {
		_ = tx.Rollback()
		return subscription.Order{}, subscription.ErrUpgradeInvalid
	}

	orders, err := tx.FindBalanceSources(builder.GetReaderID())
	if err != nil {
		return subscription.Order{}, err
	}

	wallet := subscription.NewWallet(orders, time.Now())

	builder.SetMembership(member).
		SetWallet(wallet)

	if err := builder.Build(); err != nil {
		_ = tx.Rollback()
		return subscription.Order{}, err
	}

	confirmed, err := builder.FreeUpgrade()
	if err != nil {
		_ = tx.Rollback()
		return subscription.Order{}, err
	}

	// Save upgrading schema.
	upgrade, _ := builder.UpgradeSchema()
	if err := tx.SaveUpgradeBalance(upgrade.UpgradeBalanceSchema); err != nil {
		_ = tx.Rollback()
		return subscription.Order{}, err
	}

	if err := tx.SaveProratedOrders(upgrade.Sources); err != nil {
		_ = tx.Rollback()
		return subscription.Order{}, err
	}

	// Save order
	if err := tx.SaveOrder(confirmed.Order); err != nil {
		_ = tx.Rollback()
		return subscription.Order{}, err
	}

	// Save updated membership.
	if err := tx.UpdateMember(confirmed.Membership); err != nil {
		_ = tx.Rollback()
		return subscription.Order{}, err
	}

	// Save snapshot.
	go func() {
		_ = env.BackUpMember(confirmed.Snapshot)
	}()

	if err := tx.Commit(); err != nil {
		return subscription.Order{}, err
	}

	return confirmed.Order, nil
}
