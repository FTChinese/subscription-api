package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/subs"
)

// UpgradeIntent upgrade a membership to premium for free if user's current balance
// is large enough to cover premium's price.
// This won't happen based on current restrictions of max renewal for 3 consecutive years.
// It is provided here just for logical completeness.
func (env Env) UpgradeIntent(config subs.PaymentConfig) (subs.UpgradeIntent, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.BeginOrderTx()
	if err != nil {
		sugar.Error(err)
		return subs.UpgradeIntent{}, err
	}

	member, err := tx.RetrieveMember(config.Account.MemberID())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.UpgradeIntent{}, err
	}

	orderKind, err := member.AliWxSubsKind(config.Plan.Edition)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.UpgradeIntent{}, err
	}

	// If user is not qualified to upgrade, deny it.
	if orderKind != enum.OrderKindUpgrade {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.UpgradeIntent{}, subs.ErrNotUpgradeIntent
	}

	balanceSources, err := tx.FindBalanceSources(member.MemberID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.UpgradeIntent{}, err
	}

	intent, err := config.UpgradeIntent(balanceSources, member)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.UpgradeIntent{}, err
	}

	// If this is used for preview only.
	if config.DryRun {
		if err := tx.Commit(); err != nil {
			return subs.UpgradeIntent{}, err
		}

		return intent, nil
	}

	if !intent.IsFree {
		sugar.Infof("Not a free upgrade")
		_ = tx.Rollback()
		return intent, nil
	}

	// Save order
	if err := tx.SaveOrder(intent.Result.Order); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.UpgradeIntent{}, err
	}
	sugar.Infof("Free upgrade order saved %s", intent.Result.Order)

	if err := tx.ConfirmOrder(intent.Result.Order); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.UpgradeIntent{}, err
	}

	// Save updated membership.
	if err := tx.UpdateMember(intent.Result.Membership); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.UpgradeIntent{}, err
	}

	// Save balance source as prorated.
	pos := intent.ProratedOrders(intent.Result.Order.ID)
	if err := tx.SaveProratedOrders(pos); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.UpgradeIntent{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return subs.UpgradeIntent{}, err
	}

	return intent, nil
}
