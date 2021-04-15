package account

const StmtVerifyEmailPassword = `
SELECT user_id,
	password = MD5(?) AS password_matched
FROM cmstmp01.userinfo
WHERE email = ?
LIMIT 1`

const StmtVerifyPassword = `
SELECT password = MD5(?) AS matched
FROM cmstmp01.userinfo
WHERE user_id = ?
LIMIT 1`

const StmtUpdatePassword = `
UPDATE cmstmp01.userinfo
SET password = MD5(:password),
	updated_utc = UTC_TIMESTAMP()
WHERE user_id = :ftc_id
LIMIT 1`

const StmtSignUpCount = `
SELECT COUNT(*) AS su_count
FROM user_db.client_footprint
WHERE created_utc BETWEEN CAST(? AS DATETIME) AND CAST(? AS DATETIME)
	AND source = 'signup'
GROUP BY user_ip
HAVING user_ip = INET6_ATON(?)`
