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
// Return error could be ErrLinkTargetAlreadyTaken,
// ErrLinkToExistingMember
func (env IAPEnv) Link(s apple.Subscription, id reader.MemberID) error {
	tx, err := env.BeginTx(s.Environment)
	if err != nil {
		return err
	}

	iapMember, err := tx.RetrieveAppleMember(s.OriginalTransactionID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	ftcMember, err := tx.RetrieveMember(id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	merged, err := ftcMember.MergeIAPMembership(iapMember)
	if err != nil {
		if err == subscription.ErrLinkToMultipleFTC {
			newIAPMember := s.BuildOn(iapMember)
			go func() {
				_ = env.BackUpMember(
					subscription.NewMemberSnapshot(
						iapMember,
						enum.SnapshotReasonAppleIAP,
					),
					s.Environment,
				)
			}()

			if err := tx.UpdateMember(newIAPMember); err != nil {
				_ = tx.Rollback()
			}
			if err := tx.Commit(); err != nil {
				return err
			}
			return nil
		}

		_ = tx.Rollback()
		return err
	}

	if merged.IsZero() {
		merged = s.NewMembership(id)
	} else {
		merged = s.BuildOn(merged)
	}

	// Backup current iap membership and ftc membership
	if !iapMember.IsZero() {
		// Back up
		go func() {
			_ = env.BackUpMember(
				subscription.NewMemberSnapshot(
					iapMember,
					enum.SnapshotReasonAppleIAP,
				),
				s.Environment,
			)
		}()

		if err := tx.DeleteMember(iapMember.MemberID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if !ftcMember.IsZero() {
		go func() {
			_ = env.BackUpMember(
				subscription.NewMemberSnapshot(
					ftcMember,
					enum.SnapshotReasonAppleIAP,
				),
				s.Environment,
			)
		}()

		if err := tx.DeleteMember(ftcMember.MemberID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.CreateMember(merged); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
