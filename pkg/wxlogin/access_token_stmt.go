package wxlogin

const StmtInsertAccess = `
INSERT INTO user_db.wechat_access
SET session_id = UNHEX(:session_id),
	app_id = :app_id,
	access_token = :access_token,
	expires_in = :expires_in,
	refresh_token = :refresh_token,
	open_id = :open_id,
	scope = :scope,
	union_id = :union_id,
	client_type = :platform,
	client_version = :client_version,
	user_ip = INET6_ATON(:user_ip),
	user_agent = :user_agent,
	created_utc = :created_utc,
	updated_utc = :updated_utc`

const StmtSelectAccess = `
SELECT LOWER(HEX(session_id)) AS session_id,
	app_id AS app_id,
	access_token AS access_token,
	expires_in AS expires_in,
	refresh_token AS refresh_token,
	open_id AS open_id,
	scope AS scope,
	union_id AS union_id,
	created_utc AS created_utc,
	updated_utc AS updated_utc
FROM user_db.wechat_access
WHERE session_id = UNHEX(?)
	AND app_id = ?
LIMIT 1`

const StmtUpdateAccess = `
UPDATE user_db.wechat_access
SET access_token = :access_token,
	updated_utc = :updated_utc
WHERE session_id = UNHEX(:session_id)
LIMIT 1`
