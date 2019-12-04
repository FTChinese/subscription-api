package iaprepo

import (
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/models/apple"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
)

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

	// If the ftc membership is not allowed to link to
	// iap.
	// PermitLinkApple indicates the two are not equal
	// and ftcMember is valid.
	// If iapMember is not zero, we should update it based
	// on apple transaction but do not perform the linking
	// process.
	if err := ftcMember.PermitLinkApple(iapMember); err != nil {
		if iapMember.IsZero() {
			_ = tx.Commit()
			return err
		}

		go func() {
			_ = env.BackUpMember(
				subscription.NewMemberSnapshot(
					iapMember,
					enum.SnapshotReasonAppleIAP,
				),
				s.Environment,
			)
		}()

		_ = tx.UpdateMember(s.NewMembership(iapMember.MemberID))

		_ = tx.Commit()
		return err
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

	newMember := s.NewMembership(id)
	// TODO: choose an id
	newMember.ID = iapMember.ID

	if err := tx.CreateMember(newMember); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
