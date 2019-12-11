package iaprepo

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/models/apple"
)

// Unlink deletes a membership created from IAP.
func (env IAPEnv) Unlink(s apple.Subscription) error {
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
