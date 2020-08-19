package iaprepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
)

// Link links an apple subscription to an ftc account.
// We should first retrieves membership by
// apple original transaction id and by ftc ftc separately
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
// This is a suspicious operation that should always be denied.
// Return error could be ErrTargetLinkedToOtherIAP,
// ErrHasValidNonIAPMember
// The second returned value indicates whether this is initial linking and should send an email to user.
func (env IAPEnv) Link(s apple.Subscription, id reader.MemberID) (subs.Membership, bool, error) {
	tx, err := env.BeginTx()
	if err != nil {
		return subs.Membership{}, false, err
	}

	iapMember, err := tx.RetrieveAppleMember(s.OriginalTransactionID)
	if err != nil {
		_ = tx.Rollback()
		return subs.Membership{}, false, err
	}
	ftcMember, err := tx.RetrieveMember(id)
	if err != nil {
		_ = tx.Rollback()
		return subs.Membership{}, false, err
	}

	// Merge two memberships.
	// If iap membership is already linked, the merged
	// membership won't be changed and we only need to
	// update it based on apple transaction.
	merged, err := ftcMember.MergeIAPMembership(iapMember)
	if err != nil {
		if err == subs.ErrLinkToMultipleFTC {
			newIAPMember := s.BuildOn(iapMember)
			go func() {
				_ = env.BackUpMember(
					iapMember.Snapshot(enum.SnapshotReasonAppleLink),
				)
			}()

			if e := tx.UpdateMember(newIAPMember); e != nil {
				_ = tx.Rollback()
			} else {
				_ = tx.Commit()
			}

			return subs.Membership{}, false, err
		}

		_ = tx.Rollback()
		return subs.Membership{}, false, err
	}

	if merged.IsZero() {
		merged = s.NewMembership(id)
	} else {
		merged = s.BuildOn(merged)
	}

	// Backup current iap membership and ftc membership
	if !iapMember.IsZero() {
		go func() {
			_ = env.BackUpMember(
				iapMember.Snapshot(enum.SnapshotReasonAppleLink),
			)
		}()

		if err := tx.DeleteMember(iapMember.MemberID); err != nil {
			_ = tx.Rollback()
			return subs.Membership{}, false, err
		}
	}

	if !ftcMember.IsZero() {
		go func() {
			_ = env.BackUpMember(
				ftcMember.Snapshot(enum.SnapshotReasonAppleLink),
			)
		}()

		if err := tx.DeleteMember(ftcMember.MemberID); err != nil {
			_ = tx.Rollback()
			return subs.Membership{}, false, err
		}
	}

	// Insert the merged one.
	if err := tx.CreateMember(merged); err != nil {
		_ = tx.Rollback()
		return subs.Membership{}, false, err
	}

	if err := tx.Commit(); err != nil {
		return subs.Membership{}, false, err
	}

	return merged, iapMember.IsZero(), nil
}
