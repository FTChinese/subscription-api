package ztsms

import "github.com/guregu/null"

// MobileUpdater is sued to retrieve/set user mobile
// number.
type MobileUpdater struct {
	FtcID  string      `db:"ftc_id"`
	Mobile null.String `db:"mobile_phone"`
}

const StmtLockMobileByID = `
SELECT user_id AS ftc_id,
	mobile_phone
FROM user_db.profile
WHERE user_id = ?
LIMIT 1
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
