package iaprepo

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"go.uber.org/zap"
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
// Return error could be ErrTargetLinkedToOtherIAP, ErrHasValidNonIAPMember.
func (env Env) Link(s apple.Subscription, ids reader.MemberID) (apple.LinkResult, error) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	tx, err := env.BeginTx()
	if err != nil {
		sugar.Error(err)
		return apple.LinkResult{}, err
	}

	// Try to retrieve membership by apple original transaction id.
	iapMember, err := tx.RetrieveAppleMember(s.OriginalTransactionID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return apple.LinkResult{}, err
	}
	// Try to retrieve membership by ftc id.
	ftcMember, err := tx.RetrieveMember(ids)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return apple.LinkResult{}, err
	}

	// Merge two memberships.
	// If iap membership is already linked, the merged
	// membership won't be changed and we only need to
	// update it based on apple transaction.
	ve := ftcMember.ValidateMergeIAP(iapMember)
	// If link is not allowed, there's still a possibility that IAP membership exists.
	// Caller should still needs update membership based on this subscription.
	if ve != nil {
		sugar.Error(ve)
		_ = tx.Rollback()
		return apple.LinkResult{}, ve
	}

	// If reached here, possible cases of FTC and IAP:
	// FTC	    |  IAP
	// ----------------
	// zero	    |  zero  | No backup
	// Equal    |  Equal | Backup and Update
	// Expired  |  zero  | Backup FTC and Update
	// -----------------
	// From this table we can see we only need to backup the FTC side if it exists.
	var newMmb reader.Membership
	if ftcMember.IsZero() {
		newMmb = s.NewMembership(ids)
		err := tx.CreateMember(newMmb)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return apple.LinkResult{}, err
		}

		if err := tx.Commit(); err != nil {
			sugar.Error(err)
			return apple.LinkResult{}, err
		}

		return apple.LinkResult{
			Linked:   newMmb,
			Snapshot: reader.MemberSnapshot{},
		}, nil
	}

	// The link target is not zero, but it is invalid.
	newMmb = s.BuildOn(ftcMember)
	err = tx.UpdateMember(newMmb)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return apple.LinkResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return apple.LinkResult{}, err
	}

	return apple.LinkResult{
		Linked:   newMmb,
		Snapshot: ftcMember.Snapshot(enum.SnapshotReasonAppleLink),
	}, nil
}

// Unlink deletes a membership created from IAP.
func (env Env) Unlink(originalTransID string, ids reader.MemberID) (reader.MemberSnapshot, error) {
	tx, err := env.BeginTx()

	if err != nil {
		return reader.MemberSnapshot{}, err
	}

	m, err := tx.RetrieveAppleMember(originalTransID)
	if err != nil {
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, err
	}
	if m.IsZero() {
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, sql.ErrNoRows
	}

	if m.FtcID != ids.FtcID {
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, apple.ErrUnlinkMismatchedFTC
	}

	snapshot := m.Snapshot(enum.SnapshotReasonDelete)

	if err := tx.DeleteMember(m.MemberID); err != nil {
		_ = tx.Rollback()
		return reader.MemberSnapshot{}, err
	}

	if err := tx.Commit(); err != nil {
		return reader.MemberSnapshot{}, err
	}

	return snapshot, nil
}
