package iaprepo

import (
	"database/sql"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

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
func (env Env) Link(input apple.LinkInput, account reader.FtcAccount) (apple.LinkResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.BeginTx()
	if err != nil {
		sugar.Error(err)
		return apple.LinkResult{}, err
	}

	sub, err := tx.RetrieveAppleSubs(input.OriginalTxID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return apple.LinkResult{}, err
	}

	// Possible cheating.
	if !sub.PermitLink(input.FtcID) {
		_ = tx.Rollback()
		return apple.LinkResult{}, apple.ErrIAPAlreadyLinked
	}

	// Try to retrieve membership by apple original transaction id.
	iapMember, err := tx.RetrieveAppleMember(input.OriginalTxID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return apple.LinkResult{}, err
	}
	// Try to retrieve membership by ftc id.
	ftcMember, err := tx.RetrieveMember(account.MemberID())
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return apple.LinkResult{}, err
	}

	builder := apple.LinkBuilder{
		Account:    account,
		CurrentFtc: ftcMember,
		CurrentIAP: iapMember,
		IAPSubs:    sub,
	}

	result, err := builder.Build()
	if err != nil {
		return apple.LinkResult{}, err
	}
	// If membership is take a snapshot, we must delete it.
	if !result.Snapshot.IsZero() {
		err := tx.DeleteMember(ftcMember.MemberID)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return apple.LinkResult{}, err
		}
	}
	// Save membership only when it is touched.
	if result.Touched {
		err := tx.CreateMember(result.Member)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return apple.LinkResult{}, err
		}
	}

	// Set ftc_user_id field if it is empty.
	if sub.FtcUserID.IsZero() {
		err := tx.LinkAppleSubs(input)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return apple.LinkResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return apple.LinkResult{}, err
	}

	return result, nil
}

func (env Env) ArchiveLinkCheating(link apple.LinkInput) error {
	_, err := env.db.NamedExec(apple.StmtArchiveLinkCheat, link)
	if err != nil {
		return err
	}

	return nil
}

// Unlink deletes a membership created from IAP.
// Unlike wechat unlinking FTC account which could keep the membership,
// unlinking IAP must delete the membership since IAP's owner should be unknown
// after link severed.
func (env Env) Unlink(input apple.LinkInput) (reader.MemberSnapshot, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.BeginTx()

	if err != nil {
		return reader.MemberSnapshot{}, err
	}

	// Find current membership by original transaction id.
	m, err := tx.RetrieveAppleMember(input.OriginalTxID)
	if err != nil {
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, err
	}
	// If membership is not found, stop.
	if m.IsZero() {
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, sql.ErrNoRows
	}

	// If the found membership's ftc user id does not match the requested ftc user id, stop.
	if m.FtcID.String != input.FtcID {
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, &render.ValidationError{
			Message: "IAP is not linked to the ftc account",
			Field:   "ftcId",
			Code:    render.CodeInvalid,
		}
	}

	// Delete this membership.
	if err := tx.DeleteMember(m.MemberID); err != nil {
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, err
	}

	// Try to remove ftc_user_id from apple_subscription.
	// Ignore errors.
	sub, err := tx.RetrieveAppleSubs(input.OriginalTxID)
	if err != nil {
		sugar.Error(err)
	}

	// Remove the ftc user id in apple_subscription.
	// Ignore errors.
	if sub.FtcUserID.String == input.FtcID {
		err := tx.UnlinkAppleSubs(input)
		if err != nil {
			sugar.Error(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return reader.MemberSnapshot{}, err
	}

	// Return the snapshot of the membership for archiving.
	return m.Snapshot(reader.ArchiverAppleUnlink), nil
}

func (env Env) ArchiveUnlink(link apple.LinkInput) error {
	_, err := env.db.NamedExec(apple.StmtArchiveUnlink, link)
	if err != nil {
		return err
	}

	return nil
}
