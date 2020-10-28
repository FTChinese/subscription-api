package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
)

// UpgradeIntent upgrade a membership to premium for free if user's current balance
// is large enough to cover premium's price.
// This won't happen based on current restrictions of max renewal for 3 consecutive years.
// It is provided here just for logical completeness.
func (env Env) UpgradeIntent(account reader.FtcAccount, plan product.ExpandedPlan, preview bool) (subs.UpgradeIntent, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.BeginOrderTx()
	if err != nil {
		sugar.Error(err)
		return subs.UpgradeIntent{}, err
	}

	member, err := tx.RetrieveMember(account.MemberID())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.UpgradeIntent{}, err
	}

	orderKind, err := member.AliWxSubsKind(plan.Edition)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.UpgradeIntent{}, err
	}

	// If user is not qualified to upgrade, deny it.
	if orderKind != enum.OrderKindUpgrade {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.UpgradeIntent{}, subs.ErrUpgradeInvalid
	}

	balanceSources, err := tx.FindBalanceSources(member.MemberID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.UpgradeIntent{}, err
	}

	config := subs.NewPayment(account, plan).
		WithUpgrade(preview)

	checkout := config.Checkout(balanceSources, enum.OrderKindUpgrade)

	intent, err := config.UpgradeIntent(checkout, member)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.UpgradeIntent{}, err
	}

	// If this is used for preview only.
	if preview {
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

	if err := tx.SaveProratedOrders(intent.Wallet.Sources); err != nil {
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
