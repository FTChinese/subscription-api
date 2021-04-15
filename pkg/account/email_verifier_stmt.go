package account

// StmtInsertEmailVerifier creates a record for email verification request.
// One email only has one record in db.
// If the request's email already exists, perform updates.
// Uniqueness is required since we only want to keep one
// token valid.
const StmtInsertEmailVerifier = `
INSERT INTO user_db.email_verify
SET token = UNHEX(:token),
	email = :email,
	source_url = :source_url,
	created_utc = UTC_TIMESTAMP(),
	updated_utc = UTC_TIMESTAMP()
ON DUPLICATE KEY UPDATE
	token = UNHEX(:token),
	updated_utc = UTC_TIMESTAMP()`

const StmtRetrieveEmailVerifier = `
SELECT HEX(token) AS token,
	email,
	source_url
FROM user_db.email_verify
WHERE token = UNHEX(?)
LIMIT 1`

// StmtEmailVerified set the email_verified to true.
const StmtEmailVerified = `
UPDATE cmstmp01.userinfo
	SET email_verified = TRUE
WHERE user_id = ?
LIMIT 1`
