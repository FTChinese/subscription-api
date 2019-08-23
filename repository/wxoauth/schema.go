package wxoauth

const (
	stmtInsertAccess = `
	INSERT INTO user_db.wechat_access
	SET session_id = UNHEX(?),
		app_id = ?,
		access_token = ?,
		expires_in = ?,
		refresh_token = ?,
		open_id = ?,
		scope = ?,
		union_id = ?,
		client_type = ?,
		client_version = ?,
		user_ip = INET6_ATON(?),
		user_agent = ?,
		created_utc = ?,
		updated_utc = ?`

	stmtSelectAccess = `
	SELECT LOWER(HEX(session_id)) AS sessionId, 
	    access_token AS accessToken,
		expires_in AS expiresIn,
		refresh_token AS refreshToken,
		open_id AS opendId,
		scope AS scope,
		union_id AS unionId,
		created_utc AS createdUtc,
		updated_utc AS updatedUtc
	FROM user_db.wechat_access
	WHERE session_id = UNHEX(?)
		AND app_id = ?
	LIMIT 1`

	stmtUpdateAccess = `
	UPDATE user_db.wechat_access
	SET access_token = ?,
	    updated_utc = UTC_TIMESTAMP()
	WHERE session_id = UNHEX(?)
	LIMIT 1`

	stmtInsertUserInfo = `
	INSERT INTO user_db.wechat_userinfo
	SET union_id = ?,
		nickname = ?,
		avatar_url = ?,
		gender = ?,
		country = ?,
		province = ?,
		city = ?,
		privilege = NULLIF(?, ''),
	    created_utc = UTC_TIMESTAMP(),
	    updated_utc = UTC_TIMESTAMP()
	ON DUPLICATE KEY UPDATE
		nickname = ?,
		avatar_url = ?,
		gender = ?,
		country = ?,
		province = ?,
		city = ?,
		privilege = NULLIF(?, ''),
		updated_utc = UTC_TIMESTAMP()`

	stmtUpdateUserInfo = `
	UPDATE user_db.wechat_userinfo
	SET nickname = ?,
		gender = ?,
		country = ?,
		province = ?,
		city = ?,
		avatar_url = ?,
		privilege = NULLIF(?, ''),
	    updated_utc = UTC_TIMESTAMP()
	WHERE union_id = ?`

	stmtInsertStatus = `
	INSERT INTO user_db.wechat_error_log
	SET code = ?,
		message = ?,
		created_utc = UTC_TIMESTAMP()`
)
