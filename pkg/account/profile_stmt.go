package account

const StmtProfile = `
SELECT u.user_id AS ftc_id,
	u.gender AS gender,
	u.last_name AS family_name,
	u.first_name AS given_name,
	u.birthday AS birthday,
	u.created_utc AS created_utc,
	u.updated_utc AS updated_utc,` +
	colsGetAddress + `
FROM cmstmp01.userinfo AS u
	LEFT JOIN user_db.profile AS p
		ON u.user_id = p.user_id
WHERE u.user_id = ?
LIMIT 1`

const StmtUpdateProfile = `
UPDATE cmstmp01.userinfo
SET gender = :gender,
	last_name = :family_name,
	first_name = :given_name,
	birthday = :birthday,
	updated_utc = UTC_TIMESTAMP()
WHERE user_id = :ftc_id
LIMIT 1`
