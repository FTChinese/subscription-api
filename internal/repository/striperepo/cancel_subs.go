package striperepo

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
)

// CancelSubscription cancels a subscription at period end if `cancel` is true, else reactivate it.
func (env Env) CancelSubscription(subsID string, cancel bool) (stripe.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginSubsTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	mmb, err := tx.RetrieveStripeMember(subsID)

	sugar.Infof("Current membership cancel/reactivate stripe subscription %v", mmb)

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	if !mmb.IsStripe() {
		_ = tx.Rollback()
		return stripe.SubsResult{}, errors.New("not a stripe subscription")
	}

	// If already canceled.
	if cancel && mmb.Status == enum.SubsStatusCanceled {
		_ = tx.Rollback()
		return stripe.SubsResult{
			Modified:             false,
			MissingPaymentIntent: false,
			PaymentResult:        stripe.PaymentResult{},
			Payment:              stripe.PaymentResult{},
			Subs:                 stripe.Subs{},
			Member:               mmb,
			Snapshot:             reader.MemberSnapshot{},
		}, nil
	}

	// Reactivation only applied to a canceled subscription.
	if !cancel && mmb.Status != enum.SubsStatusCanceled {
		_ = tx.Rollback()
		return stripe.SubsResult{
			Modified:             false,
			MissingPaymentIntent: false,
			PaymentResult:        stripe.PaymentResult{},
			Payment:              stripe.PaymentResult{},
			Subs:                 stripe.Subs{},
			Member:               mmb,
			Snapshot:             reader.MemberSnapshot{},
		}, nil
	}

	ss, err := env.client.CancelSubs(subsID, cancel)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

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

	sugar.Infof("Cancelled/reactivated membership %v", result.Member)

	if err := tx.UpdateMember(result.Member); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	sugar.Infof("Membership canceled/reactivated")

	return result, nil
}
