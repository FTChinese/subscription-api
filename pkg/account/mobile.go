package account

import (
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/guregu/null"
)

// MobileUpdater is sued to retrieve/set user mobile
// number in profile table.
type MobileUpdater struct {
	FtcID  string      `db:"ftc_id"`
	Mobile null.String `db:"mobile_phone"`
}

// PermitUpsertMobile verifies if dest MobileUpdater could be
// inserted/updated on profile table when it has rows of
// MobileUpdater.
// Rules to set a mobile on an account:
// The mobile must not exist yet;
// The profile table might not have row for this ftc id,
// or it has a row but mobile_phone column is empty,
// or mobile_phone column has another phone (override)
func PermitUpsertMobile(rows []MobileUpdater, dest MobileUpdater) (db.WriteKind, error) {
	rowCount := len(rows)

	// Both ftc id and mobile has a row.
	// Since they are different, ftc row has another mobile set
	// while the mobile row has another ftc id.
	if rowCount > 1 {
		return db.WriteKindDenial, ErrMobileTakenByOther
	}

	// Only one row retrieved
	if rowCount == 1 {
		current := rows[0]
		// If this row's ftc id does not match the params.FtcID,
		// it means the profile table does not have a row for this ftc id,
		// and the row is retrieved by mobile number.
		// The mobile is set on another account so this ftc id should not be allowed to use it.
		if current.FtcID != dest.FtcID {
			return db.WriteKindDenial, ErrMobileTakenByOther
		}

		// This row indeed belong to the params.FtcID.
		// Let's see its Mobile field.
		if current.Mobile.Valid {
			// This ftc id already has this mobile set.
			if current.Mobile.String == dest.Mobile.String {
				return db.WriteKindDenial, nil
			} else {
				// The ftc id has another mobile on it, you could override it.
				return db.WriteKindUpdate, nil
			}
		}

		// Mobile field is empty, fallthrough.
		return db.WriteKindUpdate, nil
	}

	return db.WriteKindInsert, nil
}

const StmtLockProfileByIDOrMobile = `
SELECT user_id AS ftc_id,
	mobile_phone
FROM user_db.profile
WHERE user_id = ? OR mobile_phone = ?
FOR UPDATE`

const StmtSetPhone = `
UPDATE user_db.profile
SET mobile_phone = :mobile_phone,
	updated_utc = UTC_TIMESTAMP()
WHERE user_id = :ftc_id
LIMIT 1`

const StmtUnsetMobile = `
UPDATE user_db.profile
SET mobile_phone = NULL
WHERE user_id = ?
LIMIT 1`
