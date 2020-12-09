package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	stripeSdk "github.com/stripe/stripe-go/v72"
)

// RefreshSubs refresh stripe subscription data if stale.
func (env Env) RefreshSubs(ss *stripeSdk.Subscription) (stripe.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginSubsTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	mmb, err := tx.RetrieveStripeMember(ss.ID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}
	sugar.Infof("Retrieved a stripe member: %+v", mmb)

	result, err := stripe.SubsBuilder{
		IDs:           mmb.MemberID,
		SS:            ss,
		CurrentMember: mmb,
		Action:        reader.ActionRefresh,
	}.Build()

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	sugar.Infof("Refreshed membership: %v", result.Member)

	// If no snapshot is taken, it indicates membership is not modified.
	if !result.Modified {
		sugar.Info("Membership not changed after refresh")
		_ = tx.Rollback()
		return result, nil
	}

	// update member
	if err := tx.UpdateMember(result.Member); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	sugar.Info("Refresh stripe subscription finished.")
	return result, nil
}
