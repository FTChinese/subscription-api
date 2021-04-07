package ztsms

const StmtSaveVerifier = `
INSERT INTO user_db.mobile_verifier
SET mobile_phone = :mobile_phone,
	sms_code = :sms_code,
	expires_in = :expires_in,
	created_utc = :created_utc,
	ftc_id = :ftc_id`

const StmtRetrieveVerifier = `
SELECT mobile_phone,
	sms_code,
	expires_in,
	created_utc,
	used_utc,
	ftc_id
FROM user_db.mobile_verifier
WHERE mobile_phone = ?
	AND sms_code = ?
	AND used_utc IS NULL
LIMIT 1`

const StmtVerifierUsed = `
UPDATE user_db.mobile_verifier
SET used_utc = :used_utc
WHERE mobile_phone = :mobile_phone
	AND sms_code = :sms_code
	AND used_utc IS NULL
LIMIT 1`

const StmtUserIDByPhone = `
SELECT user_id
FROM user_db.profile
WHERE mobile_phone = ?
LIMIT 1`

const StmtSetPhone = `
UPDATE user_db.profile
SET mobile_phone = :mobile_phone,
	updated_utc = UTC_TIMESTAMP()
WHERE user_id = :ftc_id
LIMIT 1`