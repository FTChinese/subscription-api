package iaprepo

import (
	"database/sql"
	"gitlab.com/ftchinese/subscription-api/models/apple"
)

// Unlink handles db operation to unlink iap from ftc account.
func (env IAPEnv) Unlink(s apple.Subscription) error {
	tx, err := env.BeginTx(s.Environment)

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

	if err := tx.UnlinkIAP(m.MemberID); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
