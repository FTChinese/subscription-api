package wxlogin

const (
	StmtInsertAccess = `
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

	StmtSelectAccess = `
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

	StmtUpdateAccess = `
	UPDATE user_db.wechat_access
	SET access_token = ?,
	    updated_utc = UTC_TIMESTAMP()
	WHERE session_id = UNHEX(?)
	LIMIT 1`

	StmtInsertStatus = `
	INSERT INTO user_db.wechat_error_log
	SET code = ?,
		message = ?,
		created_utc = UTC_TIMESTAMP()`
)
