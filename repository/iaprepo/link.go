package iaprepo

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
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
func (env Env) Link(s apple.Subscription, id reader.MemberID) (apple.LinkResult, error) {
	tx, err := env.BeginTx()
	if err != nil {
		return apple.LinkResult{}, err
	}

	// Try to retrieve membership by apple original transaction id.
	iapMember, err := tx.RetrieveAppleMember(s.OriginalTransactionID)
	if err != nil {
		_ = tx.Rollback()
		return apple.LinkResult{}, err
	}
	// Try to retrieve membership by ftc id.
	ftcMember, err := tx.RetrieveMember(id)
	if err != nil {
		_ = tx.Rollback()
		return apple.LinkResult{}, err
	}

	// Merge two memberships.
	// If iap membership is already linked, the merged
	// membership won't be changed and we only need to
	// update it based on apple transaction.
	ve := ftcMember.ValidateMergeIAP(iapMember)
	if ve != nil {
		// If merging is not allowed but IAP already exists, we should update it.
		if !iapMember.IsZero() {
			newIAPMember := s.BuildOn(iapMember)
			snapshot := iapMember.Snapshot(enum.SnapshotReasonAppleLink)
			go func() {
				_ = env.BackUpMember(snapshot)
			}()

			e := tx.UpdateMember(newIAPMember)
			if e != nil {
				_ = tx.Rollback()
			} else {
				_ = tx.Commit()
			}

			return apple.LinkResult{}, ve
		}

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
	var newMmb subs.Membership
	if ftcMember.IsZero() {
		newMmb = s.NewMembership(id)
	} else {
		// The merged membership is not zero, but it is invalid.
		newMmb = s.BuildOn(ftcMember)
		snapshot := ftcMember.Snapshot(enum.SnapshotReasonAppleLink)
		go func() {
			_ = env.BackUpMember(snapshot)
		}()

		err := tx.DeleteMember(ftcMember.MemberID)
		if err != nil {
			_ = tx.Rollback()
			return apple.LinkResult{}, err
		}
	}

	// Insert the merged one.
	if err := tx.CreateMember(newMmb); err != nil {
		_ = tx.Rollback()
		return apple.LinkResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return apple.LinkResult{}, err
	}

	return apple.LinkResult{
		Linked:      newMmb,
		PreviousFTC: ftcMember,
		PreviousIAP: iapMember,
	}, nil
}

// Unlink deletes a membership created from IAP.
func (env Env) Unlink(s apple.Subscription, ids reader.MemberID) error {
	tx, err := env.BeginTx()

	if err != nil {
		return err
	}

	m, err := tx.RetrieveAppleMember(s.OriginalTransactionID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if m.IsZero() {
		_ = tx.Rollback()
		return sql.ErrNoRows
	}

	if m.FtcID != ids.FtcID {
		_ = tx.Rollback()
		return apple.ErrUnlinkMismatchedFTC
	}

	snapshot := m.Snapshot(enum.SnapshotReasonDelete)

	if err := env.BackUpMember(snapshot); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.DeleteMember(m.MemberID); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
