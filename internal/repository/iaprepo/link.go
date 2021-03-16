package iaprepo

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

// GetSubAndSetFtcID retrieves an existing apple subscription by original transaction id.
// If the subscription is already have FtcUserID field set and not equal to input.FtcID,
// it indicates the subscription is claimed by other ftc account and an error will be returned.
// If the FtcUserId is empty, it will set to the input.FtcID value.
// This might result to the same ftc user id set to multiple apple subscription, which means
// one ftc account could have multiple apple subscription under it.
// In such case, only one of the subscriptions will be reflected in a user's membership.
func (env Env) GetSubAndSetFtcID(input apple.LinkInput) (apple.Subscription, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.BeginTx()
	if err != nil {
		sugar.Error(err)
		return apple.Subscription{}, err
	}

	sub, err := tx.RetrieveAppleSubs(input.OriginalTxID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return apple.Subscription{}, err
	}

	// If sub.FtcUserID is not empty, and not equal to input.FtcID
	// link should be denied and this is possible cheating.
	if !sub.PermitLink(input.FtcID) {
		sugar.Infof("Link %s to %s is not permitted", input.FtcID, input.OriginalTxID)
		_ = tx.Rollback()
		return apple.Subscription{}, apple.ErrIAPAlreadyLinked
	}

	// Set ftc_user_id field if it is empty.
	if sub.FtcUserID.IsZero() {
		sub.FtcUserID = null.StringFrom(input.FtcID)
		err := tx.LinkAppleSubs(input)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return apple.Subscription{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return apple.Subscription{}, err
	}

	return sub, nil
}

// Link links an apple subscription to an ftc account.
// We should first retrieves membership by
// apple original transaction id and by ftc id separately
// to see if the two sides have existing records.
// We need to pay special attention to those two case:
//
// 1. User changes apple ID while trying to link to the same
// ftc account:
//
// 	Apple ID A + FTC ID A
// 	Apple ID B + FTC ID A
//
// 2. One apple id is trying to link to multiple ftc id:
//
//	Apple ID A + FTC ID A
//	Apple ID A + FTC ID B
//
// This is a suspicious operation that should always be denied.
// The returned error could be *render.ValidationError
// if link if forbidden.
func (env Env) Link(result apple.LinkResult) error {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.BeginTx()
	if err != nil {
		sugar.Error(err)
		return err
	}

	// Save membership only when it is touched.
	// If membership is take a snapshot, we must delete it.
	if !result.Snapshot.IsZero() {
		// Should we lock this row first?
		err := tx.DeleteMember(result.Snapshot.UserIDs)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return err
		}
	}

	err = tx.CreateMember(result.Member)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	sugar.Info("Link finished")

	return nil
}

func (env Env) ArchiveLinkCheating(link apple.LinkInput) error {
	_, err := env.dbs.Write.NamedExec(apple.StmtArchiveLinkCheat, link)
	if err != nil {
		return err
	}

	return nil
}

// Unlink removes FtcUserID from a Subscription, and then delete the
// membership if this subscription is currently being used as user's default membership.
func (env Env) Unlink(input apple.LinkInput) (apple.UnlinkResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.BeginTx()

	if err != nil {
		return apple.UnlinkResult{}, err
	}

	// Try to remove ftc_user_id from apple_subscription.
	// Ignore errors.
	sub, err := tx.RetrieveAppleSubs(input.OriginalTxID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return apple.UnlinkResult{}, err
	}

	// Remove the ftc user id in apple_subscription.
	// Ignore errors.
	if sub.FtcUserID.IsZero() || sub.FtcUserID.String != input.FtcID {
		_ = tx.Rollback()
		return apple.UnlinkResult{}, &render.ValidationError{
			Message: "IAP is not linked to the ftc account",
			Field:   "ftcId",
			Code:    render.CodeInvalid,
		}
	}

	err = tx.UnlinkAppleSubs(input)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return apple.UnlinkResult{}, err
	}

	// Find current membership by original transaction id.
	m, err := tx.RetrieveAppleMember(input.OriginalTxID)
	if err != nil {
		_ = tx.Rollback()
		return apple.UnlinkResult{}, err
	}

	// If membership is not found, stop and commit previous operations.
	if m.IsZero() {
		if err := tx.Commit(); err != nil {
			return apple.UnlinkResult{}, err
		}
		return apple.UnlinkResult{
			IAPSubs:  sub,
			Snapshot: reader.MemberSnapshot{},
		}, nil
	}

	// If the found membership's ftc user id does not match the requested ftc user id, stop.
	if m.FtcID.String != input.FtcID {
		_ = tx.Rollback()
		return apple.UnlinkResult{}, &render.ValidationError{
			Message: "IAP is not linked to the ftc account",
			Field:   "ftcId",
			Code:    render.CodeInvalid,
		}
	}

	// Delete this membership.
	if err := tx.DeleteMember(m.UserIDs); err != nil {
		_ = tx.Rollback()
		return apple.UnlinkResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return apple.UnlinkResult{}, err
	}

	// Return the snapshot of the membership for archiving.
	return apple.UnlinkResult{
		IAPSubs:  sub,
		Snapshot: m.Snapshot(reader.ArchiverAppleUnlink),
	}, nil
}

func (env Env) ArchiveUnlink(link apple.LinkInput) error {
	_, err := env.dbs.Write.NamedExec(apple.StmtArchiveUnlink, link)
	if err != nil {
		return err
	}

	return nil
}
