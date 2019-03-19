package model

import (
	"github.com/FTChinese/go-rest"
	"gitlab.com/ftchinese/subscription-api/wxlogin"
)

// SaveWxAccess saves the access token related data after acquired from wechat api.
func (env Env) SaveWxAccess(appID string, acc wxlogin.OAuthAccess, c gorest.ClientApp) error {
	query := `
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

	_, err := env.db.Exec(query,
		acc.SessionID,
		appID,
		acc.AccessToken,
		acc.ExpiresIn,
		acc.RefreshToken,
		acc.OpenID,
		acc.Scope,
		acc.UnionID,
		c.ClientType,
		c.Version,
		c.UserIP,
		c.UserAgent,
		acc.CreatedAt,
		acc.UpdatedAt,
	)

	if err != nil {
		logger.WithField("trace", "SaveAccess").Error(err)
		return err
	}

	return nil
}

// LoadWxAccess retrieves previously saved access token by open id.
// Is it possbile wechat generate different openID under the same app?
// Or is it possible wechat generate same openID for different app?
// What if iOS and Android used the same Wechat app? Will they use the same access token or different one?
// Open ID is always the same the under the same Wechat app.
func (env Env) LoadWxAccess(appID, sessionID string) (wxlogin.OAuthAccess, error) {
	query := `
	SELECT access_token AS accessToken,
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

	var acc wxlogin.OAuthAccess
	err := env.db.QueryRow(query, sessionID, appID).Scan(
		&acc.AccessToken,
		&acc.ExpiresIn,
		&acc.RefreshToken,
		&acc.OpenID,
		&acc.Scope,
		&acc.UnionID,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)

	if err != nil {
		logger.WithField("trace", "LoadAccess").Error(err)
		return acc, err
	}

	return acc, nil
}

// UpdateWxAccess saves refreshed access token.
func (env Env) UpdateWxAccess(sessionID, accessToken string) error {
	query := `
	UPDATE user_db.wechat_access
	SET access_token = ?,
	    updated_utc = UTC_TIMESTAMP()
	WHERE session_id = UNHEX(?)
	LIMIT 1`

	_, err := env.db.Exec(query,
		accessToken,
		sessionID,
	)

	if err != nil {
		logger.WithField("trace", "UpdateAccess").Error(err)
		return err
	}

	return nil
}

// SaveWxUser from wechat API.
// Since a user can authorize multiple times, use ON DUPLICATE to handle unique key constraint.
func (env Env) SaveWxUser(u wxlogin.UserInfo) error {
	query := `
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

	_, err := env.db.Exec(query,
		u.UnionID,
		u.NickName,
		u.AvatarURL,
		u.GetGender(),
		u.Country,
		u.Province,
		u.City,
		u.GetPrivilege(),
		u.NickName,
		u.AvatarURL,
		u.GetGender(),
		u.Country,
		u.Province,
		u.City,
		u.GetPrivilege(),
	)

	if err != nil {
		logger.WithField("trace", "SaveUserInfo").Error(err)
		return err
	}

	return nil
}

// UpdateWxUser update data of one union id.
func (env Env) UpdateWxUser(u wxlogin.UserInfo) error {
	query := `
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

	_, err := env.db.Exec(query,
		u.NickName,
		u.GetGender(),
		u.Country,
		u.Province,
		u.City,
		u.AvatarURL,
		u.GetPrivilege(),
		u.UnionID,
	)

	if err != nil {
		logger.WithField("trace", "UpdateUserInfo").Error(err)
		return err
	}

	return nil
}

// SaveWxStatus saves Wechat OAuth API error response into data so that we could know what errors on hell Wechat actually produced.
func (env Env) SaveWxStatus(code int64, message string) error {
	query := `
	INSERT INTO user_db.wechat_error_log
	SET code = ?,
		message = ?,
		created_utc = UTC_TIMESTAMP()`

	_, err := env.db.Exec(query,
		code,
		message,
	)

	if err != nil {
		logger.WithField("trace", "SaveWxError").Error(err)
		return err
	}

	return nil
}
