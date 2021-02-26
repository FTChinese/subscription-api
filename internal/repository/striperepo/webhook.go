package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	stripeSdk "github.com/stripe/stripe-go/v72"
)

// OnSubscription save stripe subscription and optionally update membership linked to it.
func (env Env) OnSubscription(ss *stripeSdk.Subscription, param stripe.SubsResultParams) (stripe.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar().
		With("webhook", "stripe-subscription").
		With("id", ss.ID)

	tx, err := env.beginSubsTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	// Retrieve current membership by ftc id.
	// If current membership is empty, we should create it.
	currMmb, err := tx.RetrieveMember(param.UserIDs)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	param.CurrentMember = currMmb

	result, err := stripe.NewSubsResult(ss, param)

	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	// Ensure that current membership is create via stripe.
	if !result.Subs.ShouldUpsert(currMmb) {
		_ = tx.Rollback()
		sugar.Infof("Stripe subscription cannot update/insert its membership")
		return stripe.SubsResult{
			Modified:             false,
			MissingPaymentIntent: false,
			Subs:                 result.Subs,
			Member:               currMmb,
			Snapshot:             reader.MemberSnapshot{},
		}, nil
	}

	// If nothing changed.
	if !result.Member.IsModified(currMmb) {
		_ = tx.Rollback()
		return stripe.SubsResult{
			Modified:             false,
			MissingPaymentIntent: false,
			Subs:                 result.Subs,
			Member:               result.Member,
			Snapshot:             reader.MemberSnapshot{},
		}, nil
	}

	// Insert to update membership.
	if currMmb.IsZero() {
		if err := tx.CreateMember(result.Member); err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsResult{}, err
		}
	} else {
		if err := tx.UpdateMember(result.Member); err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return stripe.SubsResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	return result, nil
}
