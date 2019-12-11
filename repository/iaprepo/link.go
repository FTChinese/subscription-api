package iaprepo

import (
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/models/apple"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
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
func (env IAPEnv) Link(s apple.Subscription, id reader.MemberID) (subscription.Membership, error) {
	tx, err := env.BeginTx()
	if err != nil {
		return subscription.Membership{}, err
	}

	iapMember, err := tx.RetrieveAppleMember(s.OriginalTransactionID)
	if err != nil {
		_ = tx.Rollback()
		return subscription.Membership{}, err
	}
	ftcMember, err := tx.RetrieveMember(id)
	if err != nil {
		_ = tx.Rollback()
		return subscription.Membership{}, err
	}

	// Merge two memberships.
	// If iap membership is already linked, the merged
	// membership won't be changed and we only need to
	// update it based on apple transaction.
	merged, err := ftcMember.MergeIAPMembership(iapMember)
	if err != nil {
		if err == subscription.ErrLinkToMultipleFTC {
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

			return subscription.Membership{}, err
		}

		_ = tx.Rollback()
		return subscription.Membership{}, err
	}

	if merged.IsZero() {
		merged = s.NewMembership(id)
	} else {
		merged = s.BuildOn(merged)
	}

	// Backup current iap membership and ftc membership
	if !iapMember.IsZero() {
		// Back up
		iapMember.GenerateID()
		go func() {
			_ = env.BackUpMember(
				iapMember.Snapshot(enum.SnapshotReasonAppleLink),
			)
		}()

		if err := tx.DeleteMember(iapMember.MemberID); err != nil {
			_ = tx.Rollback()
			return subscription.Membership{}, err
		}
	}

	if !ftcMember.IsZero() {
		ftcMember.GenerateID()
		go func() {
			_ = env.BackUpMember(
				ftcMember.Snapshot(enum.SnapshotReasonAppleLink),
			)
		}()

		if err := tx.DeleteMember(ftcMember.MemberID); err != nil {
			_ = tx.Rollback()
			return subscription.Membership{}, err
		}
	}

	// Insert the merged one.
	if err := tx.CreateMember(merged); err != nil {
		_ = tx.Rollback()
		return subscription.Membership{}, err
	}

	if err := tx.Commit(); err != nil {
		return subscription.Membership{}, err
	}

	return merged, nil
}
