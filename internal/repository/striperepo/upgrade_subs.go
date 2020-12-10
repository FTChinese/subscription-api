package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
)

// UpgradeSubscription switches subscription plan.
func (env Env) UpgradeSubscription(cfg stripe.PaymentConfig) (stripe.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginSubsTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	// Retrieve current membership.
	mmb, err := tx.RetrieveMember(cfg.Account.MemberID())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, nil
	}

	err = mmb.PermitStripeUpgrade(cfg.Plan.Edition)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	ss, err := env.client.UpgradeSubs(mmb.StripeSubsID.String, cfg)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}
	sugar.Infof("Subscription id %s, status %s, payment intent status %s", ss.ID, ss.Status, ss.LatestInvoice.PaymentIntent.Status)

	result, err := stripe.SubsBuilder{
		IDs:           mmb.MemberID,
		SS:            ss,
		CurrentMember: mmb,
		Action:        reader.ActionUpgrade,
	}.Build()

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	sugar.Infof("Upgraded membership %v", result.Member)

	if err := tx.UpdateMember(result.Member); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	return result, nil
}
