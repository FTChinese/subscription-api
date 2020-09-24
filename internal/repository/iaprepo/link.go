package iaprepo

import (
	"database/sql"
	"errors"

	"github.com/FTChinese/go-rest/enum"
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
func (env Env) Link(account reader.FtcAccount, iapSubs apple.Subscription) (apple.LinkResult, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	tx, err := env.BeginTx()
	if err != nil {
		sugar.Error(err)
		return apple.LinkResult{}, err
	}

	// Try to retrieve membership by apple original transaction id.
	iapMember, err := tx.RetrieveAppleMember(iapSubs.OriginalTransactionID)
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

	// Merge two memberships.
	// If iap membership is already linked, the merged
	// membership won't be changed and we only need to
	// update it based on apple transaction.
	err = ftcMember.ValidateMergeIAP(iapMember, iapSubs)
	if err != nil {
		sugar.Error(err)

		if errors.Is(err, reader.ErrIAPFtcLinked) {
			_ = tx.Rollback()
			return apple.LinkResult{
				Initial:  false,
				Linked:   iapMember,
				Snapshot: reader.MemberSnapshot{},
			}, nil
		}

		_ = tx.Rollback()
		return apple.LinkResult{}, err
	}

	// If reached here, possible cases of FTC and IAP:
	// FTC	    |  IAP
	// ----------------
	// zero	    |  zero  | No backup
	// Equal    |  Equal | Stop
	// Expired  |  zero  | Backup FTC and Update
	// -----------------
	// From this table we can see we only need to backup the FTC side if it exists.
	var newMmb = iapSubs.NewMembership(account.MemberID())

	if !ftcMember.IsZero() {
		err := tx.DeleteMember(ftcMember.MemberID)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return apple.LinkResult{}, err
		}
	}

	err = tx.CreateMember(newMmb)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return apple.LinkResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return apple.LinkResult{}, err
	}

	return apple.LinkResult{
		Initial:  iapMember.IsZero(), // As long as iap side is zero, this is initial link.
		Linked:   newMmb,
		Snapshot: ftcMember.Snapshot(enum.SnapshotReasonAppleLink),
	}, nil
}

// Unlink deletes a membership created from IAP.
// Unlike wechat unlinking FTC account which could keep the membership,
// unlinking IAP must delete the membership since IAP's owner should be unknown
// after link severed.
func (env Env) Unlink(input apple.LinkInput) (reader.MemberSnapshot, error) {
	tx, err := env.BeginTx()

	if err != nil {
		return reader.MemberSnapshot{}, err
	}

	m, err := tx.RetrieveAppleMember(input.OriginalTxID)
	if err != nil {
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, err
	}
	if m.IsZero() {
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, sql.ErrNoRows
	}

	if m.FtcID.String != input.FtcID {
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, &render.ValidationError{
			Message: "IAP is not linked to the ftc account",
			Field:   "ftcId",
			Code:    render.CodeInvalid,
		}
	}

	if err := tx.DeleteMember(m.MemberID); err != nil {
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, err
	}

	if err := tx.Commit(); err != nil {
		return reader.MemberSnapshot{}, err
	}

	return m.Snapshot(enum.SnapshotReasonAppleUnlink), nil
}
