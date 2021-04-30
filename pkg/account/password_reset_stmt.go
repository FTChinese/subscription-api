package account

const StmtInsertPwResetSession = `
INSERT INTO user_db.password_reset
SET email = :email,
	source_url = :source_url,
	token = UNHEX(:token),
	app_code = :app_code,
	expires_in = :expires_in,
	created_utc = UTC_TIMESTAMP()`

// Do not removed the time comparison condition.
// It could reduce the chance of collision for app_code.
const selectPwResetSession = `
SELECT email, 
	source_url,
	LOWER(HEX(token)) AS token,
	app_code,
	is_used,
	expires_in,
	created_utc
FROM user_db.password_reset
WHERE is_used = 0
	AND DATE_ADD(created_utc, INTERVAL expires_in SECOND) > UTC_TIMESTAMP()
`

// StmtPwResetSessionByToken retrieves a password reset session
// by token for web app.
const StmtPwResetSessionByToken = selectPwResetSession + `
	AND token = UNHEX(?)
LIMIT 1`

// StmtPwResetSessionByCode retrieves a password reset session
// by email + appCode on mobile apps.
const StmtPwResetSessionByCode = selectPwResetSession + `
	AND app_code = ?
	AND email = ?
LIMIT 1`

// StmtDisablePwResetToken flags a password reset token as invalid.
const StmtDisablePwResetToken = `
UPDATE user_db.password_reset
	SET is_used = 1
WHERE token = UNHEX(?)
LIMIT 1`
