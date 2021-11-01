package account

// StmtVerifyEmailPassword is used to authenticate an email
// and password when logging in.
const StmtVerifyEmailPassword = `
SELECT user_id,
	password = MD5(?) AS password_matched
FROM cmstmp01.userinfo
WHERE email = ?
LIMIT 1`

// StmtVerifyPassword is used to verify a user's existing password.
const StmtVerifyPassword = `
SELECT password = MD5(?) AS matched
FROM cmstmp01.userinfo
WHERE user_id = ?
LIMIT 1`

// StmtUpdatePassword changes a user's password.
// Map to IDCredentials.
const StmtUpdatePassword = `
UPDATE cmstmp01.userinfo
SET password = MD5(:password),
	updated_utc = UTC_TIMESTAMP()
WHERE user_id = :ftc_id
LIMIT 1`
