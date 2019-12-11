package subrepo

import "gitlab.com/ftchinese/subscription-api/models/subscription"

func (env SubEnv) FreeUpgrade(builder *subscription.OrderBuilder) (subscription.Order, error) {

	otx, err := env.BeginOrderTx()
	if err != nil {
		return subscription.Order{}, err
	}

	// The same process as preview upgrading balance.
	if err := otx.PreviewUpgrade(builder); err != nil {
		_ = otx.Rollback()
		return subscription.Order{}, err
	}

	if err := builder.Build(); err != nil {
		_ = otx.Rollback()
		return subscription.Order{}, err
	}

	confirmed, err := builder.FreeUpgrade()
	if err != nil {
		_ = otx.Rollback()
		return subscription.Order{}, err
	}

	// Save upgrading schema.
	upgrade, _ := builder.UpgradeSchema()
	if err := otx.SaveUpgradeBalance(upgrade.UpgradeBalanceSchema); err != nil {
		_ = otx.Rollback()
		return subscription.Order{}, err
	}

	if err := otx.SaveProratedOrders(upgrade.Sources); err != nil {
		_ = otx.Rollback()
		return subscription.Order{}, err
	}

	// Save order
	if err := otx.SaveOrder(confirmed.Order); err != nil {
		_ = otx.Rollback()
		return subscription.Order{}, err
	}

	// Save updated membership.
	if err := otx.UpdateMember(confirmed.Membership); err != nil {
		_ = otx.Rollback()
		return subscription.Order{}, err
	}

	// Save snapshot.
	go func() {
		_ = env.BackUpMember(confirmed.Snapshot)
	}()

	if err := otx.Commit(); err != nil {
		return subscription.Order{}, err
	}

	return confirmed.Order, nil
}
