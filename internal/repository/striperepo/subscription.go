package striperepo

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	stripePkg "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/guregu/null"
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
func (env Env) CreateSubscription(input stripePkg.SubsInput) (stripePkg.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginOrderTx()
	if err != nil {
		sugar.Error(err)
		return stripePkg.SubsResult{}, err
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
		return stripePkg.SubsResult{}, err
	}

	// Check whether creating stripe subscription is allowed for this member.
	subsKind, ve := mmb.StripeSubsKind(input.Edition)
	if ve != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripePkg.SubsResult{}, err
	}
	if subsKind != enum.OrderKindCreate {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripePkg.SubsResult{}, errors.New("invalid request to create Stripe subscription")
	}

	sugar.Info("Creating stripe subscription")
	// Contact Stripe API.
	ss, err := input.CreateSubs()
	sugar.Infof("Subscription id %s, status %s, payment intent status %s", ss.ID, ss.Status, ss.LatestInvoice.PaymentIntent.Status)

	// {"status":400,
	// "message":"Keys for idempotent requests can only be used with the same parameters they were first used with. Try using a key other than '4a857eb3-396c-4c91-a8f1-4014868a8437' if you meant to execute a different request.","request_id":"req_Dv6N7d9lF8uDHJ",
	// "type":"idempotency_error"
	// }
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripePkg.SubsResult{}, err
	}

	// Create Membership based on stripe subscription.
	// Keep existing membership's union id if exists.
	newMmb := input.NewMembership(mmb, ss)
	sugar.Infof("A new stripe membership: %+v", newMmb)

	// Create membership from Stripe subscription.
	if !mmb.IsZero() {
		if err := tx.DeleteMember(mmb.MemberID); err != nil {
			_ = tx.Rollback()
			return stripePkg.SubsResult{}, err
		}
	}

	if err := tx.CreateMember(newMmb); err != nil {
		_ = tx.Rollback()
		return stripePkg.SubsResult{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripePkg.SubsResult{}, err
	}

	return stripePkg.SubsResult{
		StripeSubs: ss,
		Member:     newMmb,
		Snapshot:   mmb.Snapshot(reader.ArchiverStripeCreate),
	}, err
}

// SaveSubsError saves any error in stripe response.
func (env Env) SaveSubsError(e stripePkg.APIError) error {
	_, err := env.db.NamedExec(stripePkg.StmtSaveAPIError, e)

	if err != nil {
		return err
	}

	return nil
}

// GetSubscription refresh stripe subscription data if stale.
func (env Env) GetSubscription(ftcID string) (stripePkg.SubsResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.beginOrderTx()
	if err != nil {
		sugar.Error(err)
		return stripePkg.SubsResult{}, err
	}

	mmb, err := tx.RetrieveMember(reader.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(ftcID),
		UnionID:    null.String{},
	}.MustNormalize())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripePkg.SubsResult{}, err
	}

	if !mmb.IsStripe() {
		sugar.Infof("Stripe membership not found for %v", ftcID)
		_ = tx.Rollback()
		return stripePkg.SubsResult{}, sql.ErrNoRows
	}
	sugar.Infof("Retrieve a member: %+v", mmb)

	ss, err := stripePkg.GetSubscription(mmb.StripeSubsID.String)

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripePkg.SubsResult{}, err
	}

	sugar.Infof("Subscription id %s, status %s", ss.ID, ss.Status)

	newMmb := stripePkg.RefreshMembership(mmb, ss)

	sugar.Infof("Refreshed membership: %+v", newMmb)

	// update member
	if err := tx.UpdateMember(newMmb); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return stripePkg.SubsResult{}, err
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return stripePkg.SubsResult{}, err
	}

	sugar.Info("Refresh stripe subscription finished.")
	return stripePkg.SubsResult{
		StripeSubs: ss,
		Member:     newMmb,
		Snapshot:   reader.MemberSnapshot{},
	}, nil
}
