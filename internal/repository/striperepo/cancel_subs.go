package striperepo

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/guregu/null"
)

// CancelSubscription cancels a subscription at period end if `CancelParams.Cancel` is true, else reactivate it.
// Here the cancel actually does not delete the subscription.
// It only indicates this subscription won't be automatically renews at period end.
// A canceled subscription is still in active state.
// When stripe says the status is canceled, it means the subscription is expired, deleted, and it won't charge upon period ends.
func (env Env) CancelSubscription(params stripe.CancelParams) (stripe.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginStripeTx()
	if err != nil {
		sugar.Error(err)
		return stripe.SubsResult{}, err
	}

	mmb, err := tx.RetrieveMember(pkg.UserIDs{
		CompoundID: params.FtcID,
		FtcID:      null.StringFrom(params.FtcID),
		UnionID:    null.String{},
	})
	sugar.Infof("Current membership cancel/reactivate stripe subscription %v", mmb)

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	if !mmb.IsStripe() {
		_ = tx.Rollback()
		return stripe.SubsResult{}, sql.ErrNoRows
	}

	if mmb.StripeSubsID.String != params.SubID {
		_ = tx.Rollback()
		return stripe.SubsResult{}, sql.ErrNoRows
	}

	// If you want to cancel it, and membership is not auto renewal,
	// it means it is already canceled.
	// If cancel is false, you are reactivating a canceled subscription.
	// If the membership is not auto renewal, it means the member
	// is already reactivated, or not canceled at all.
	// Only cancel and auto renewal are consistent should you proceed.
	if params.Cancel != mmb.AutoRenewal {
		_ = tx.Rollback()
		return stripe.SubsResult{
			Modified:             false,
			MissingPaymentIntent: false,
			Subs:                 stripe.Subs{},
			Member:               mmb,
			Snapshot:             reader.MemberSnapshot{},
		}, nil
	}

	ss, err := env.client.CancelSubs(params.SubID, params.Cancel)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripe.SubsResult{}, err
	}

	sugar.Infof("Canceled/reactivated subscription %s, status %s", ss.ID, ss.Status)

	var action reader.ArchiveAction
	if params.Cancel {
		action = reader.ActionCancel
	} else {
		action = reader.ActionReactivate
	}

	result, err := stripe.NewSubsResult(ss, stripe.SubsResultParams{
		UserIDs:       mmb.UserIDs,
		CurrentMember: mmb,
		Action:        action,
	})

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
