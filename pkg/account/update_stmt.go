package account

const StmtUpdateEmail = `
UPDATE cmstmp01.userinfo
	SET email = :email,
		email_verified = 0,
		updated_utc = UTC_TIMESTAMP()
WHERE user_id = :ftc_id
LIMIT 1`

const StmtBackUpEmail = `
INSERT INTO cmstmp01.email_history
SET	user_id = :ftc_id,
	user_name = IFNULL(:user_name, ''),
	email = :email,
	modify_date = UTC_TIMESTAMP()`

const StmtUpdateUserName = `
UPDATE cmstmp01.userinfo
SET user_name = :user_name,
	updated_utc = UTC_TIMESTAMP()
WHERE user_id = :ftc_id
LIMIT 1`
