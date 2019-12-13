package access

const stmtOAuth = `
SELECT access_token,
	is_active,
	expires_in,
	created_utc
FROM oauth.access
WHERE access_token = UNHEX(?)
LIMIT 1`
