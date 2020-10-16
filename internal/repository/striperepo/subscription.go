package striperepo

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
)

// CreateSubscription creates a new subscription.
// If performs multiple steps:
// 1. Find existing membership if exist;
// 2. Check what kind of subscription it is, and whether we allow it to continue;
// 3. Create subscription at Stripe API.
// 4. Build FTC membership from Stripe subscription.
// It returns stripe's subscription as is.
//
// error could be:
// util.ErrNonStripeValidSub
// util.ErrActiveStripeSub
// util.ErrUnknownSubState
func (env Env) CreateSubscription(input ftcStripe.SubsInput) (*stripe.Subscription, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginOrderTx()
	if err != nil {
		sugar.Error(err)
		return nil, err
	}

	// Retrieve member for this user to check whether the operation is allowed.
	mmb, err := tx.RetrieveMember(reader.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(input.FtcID),
		UnionID:    null.String{},
	}.MustNormalize())
	sugar.Infof("A stripe membership: %+v", mmb)

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return nil, err
	}

	// Check whether creating stripe subscription is allowed for this member.
	subsKind, ve := mmb.StripeSubsKind(input.Edition)
	if ve != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return nil, err
	}
	if subsKind != enum.OrderKindCreate {
		sugar.Error(err)
		_ = tx.Rollback()
		return nil, errors.New("invalid request to create Stripe subscription")
	}

	sugar.Info("Creating stripe subscription")
	// Contact Stripe API.
	ss, err := input.CreateSubs()

	// {"status":400,
	// "message":"Keys for idempotent requests can only be used with the same parameters they were first used with. Try using a key other than '4a857eb3-396c-4c91-a8f1-4014868a8437' if you meant to execute a different request.","request_id":"req_Dv6N7d9lF8uDHJ",
	// "type":"idempotency_error"
	// }
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return nil, err
	}

	// Create membership from Stripe subscription.
	if mmb.IsZero() {
		newMmb := input.NewMembership(ss)
		sugar.Infof("A new stripe membership: %+v", newMmb)
		if err := tx.CreateMember(newMmb); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	} else {
		newMmb := input.UpdateMembership(mmb, ss)
		sugar.Infof("Updated stripe membership: %+v", newMmb)
		if err := tx.UpdateMember(newMmb); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return nil, err
	}

	return ss, err
}

// SaveSubsError saves any error in stripe response.
func (env Env) SaveSubsError(e ftcStripe.APIError) error {
	_, err := env.db.NamedExec(ftcStripe.StmtSaveAPIError, e)

	if err != nil {
		return err
	}

	return nil
}

// GetSubscription refresh stripe subscription data if stale.
func (env Env) GetSubscription(ftcID string) (*stripe.Subscription, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginOrderTx()
	if err != nil {
		sugar.Error(err)
		return nil, err
	}

	mmb, err := tx.RetrieveMember(reader.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(ftcID),
		UnionID:    null.String{},
	}.MustNormalize())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return nil, err
	}

	if mmb.IsZero() {
		sugar.Infof("Membership not found for %v", ftcID)
		_ = tx.Rollback()
		return nil, sql.ErrNoRows
	}

	sugar.Infof("Retrieve a member: %+v", mmb)

	// If this membership is not a stripe subscription, deny further actions
	if mmb.PaymentMethod != enum.PayMethodStripe {
		sugar.Error(err)
		_ = tx.Rollback()
		return nil, sql.ErrNoRows
	}
	if mmb.StripeSubsID.IsZero() {
		sugar.Error(err)
		_ = tx.Rollback()
		return nil, sql.ErrNoRows
	}

	ss, err := ftcStripe.GetSubscription(mmb.StripeSubsID.String)

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return nil, err
	}

	newMmb := ftcStripe.RefreshMembership(mmb, ss)

	sugar.Infof("Refreshed membership: %+v", newMmb)

	// update member
	if err := tx.UpdateMember(newMmb); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return nil, err
	}

	sugar.Info("Refreshed finished.")
	return ss, nil
}
