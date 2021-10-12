package account

import "github.com/guregu/null"

// MobileUpdater is sued to retrieve/set user mobile
// number in profile table.
type MobileUpdater struct {
	FtcID  string      `db:"ftc_id"`
	Mobile null.String `db:"mobile_phone"`
}

// IsMobileSettable verifies if dest MobileUpdater could be
// inserted/updated on profile table when it has rows of
// MobileUpdater.
func IsMobileSettable(rows []MobileUpdater, dest MobileUpdater) error {
	rowCount := len(rows)

	if rowCount > 1 {
		return ErrMobileTaken
	}

	if rowCount == 1 {
		current := rows[0]
		// If this row's ftc id does not match the params.FtcID,
		// it means this row is retrieve by mobile number and
		// the mobile is set on another account.
		if current.FtcID != dest.FtcID {
			return ErrMobileTaken
		}

		// This row indeed belong to the params.FtcID.
		// Let's see its Mobile field.
		// Only allowed to proceed if the Mobile is empty.
		if current.Mobile.Valid {
			// This ftc id already has this mobile set.
			if current.Mobile.String == dest.Mobile.String {
				return ErrMobileSet
			} else {
				return ErrAccountHasMobileSet
			}
		}
		// Mobile field is empty, fallthrough.
	}

	return nil
}

const StmtLockProfileByIDOrMobile = `
SELECT user_id AS ftc_id,
	mobile_phone
FROM user_db.profile
WHERE user_id = ? OR mobile_phone = ?
FOR UPDATE`

const colsSetPhone = `
mobile_phone = :mobile_phone,
updated_utc = UTC_TIMESTAMP()
`

// StmtUpsertPhone set a mobile phone to user account.
const StmtUpsertPhone = `
INSERT INTO user_db.profile
SET user_id = :ftc_id,
` + colsSetPhone + `
ON DUPLICATE KEY UPDATE
` + colsSetPhone

const StmtUnsetMobile = `
UPDATE user_db.profile
SET mobile_phone = NULL
WHERE user_id = ?
LIMIT 1`
