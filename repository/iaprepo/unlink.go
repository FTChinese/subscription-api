package iaprepo

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

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
