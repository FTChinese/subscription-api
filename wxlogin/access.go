package wxlogin

import (
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/util"
)

// OAuthAccess contains data returned by exchange access token with oauth code.
type OAuthAccess struct {
	AccessToken  string      `json:"access_token"`
	ExpiresIn    int64       `json:"expires_in"`
	RefreshToken string      `json:"refresh_token"`
	OpenID       string      `json:"openid"`
	Scope        string      `json:"scope"`
	UnionID      null.String `json:"unionid"`
}

// SaveAccess saves the access token related data after acquired from wechat api.
// Or re-authorize if refresh token expired.
func (env Env) SaveAccess(acc OAuthAccess, c util.RequestClient) error {
	query := `INSERT INTO user_db.wechat_access
	SET access_token = ?,
		expires_in = ?,
		refresh_token = ?,
		open_id = ?,
		scope = ?,
		union_id = ?,
		client_type = ?,
		client_version = NULLIF(?, ''),
		user_ip = INET6_ATON(?)
	ON DUPLICATE KEY UPDATE access_token = ?,
		expires_in = ?,
		refresh_token = ?,
		client_type = ?,
		client_version = NULLIF(?, ''),
		user_ip = INET6_ATON(?)`

	_, err := env.DB.Exec(query,
		acc.AccessToken,
		acc.ExpiresIn,
		acc.RefreshToken,
		acc.OpenID,
		acc.Scope,
		acc.UnionID,
		c.ClientType,
		c.Version,
		c.UserIP,
		acc.AccessToken,
		acc.ExpiresIn,
		acc.RefreshToken,
		c.ClientType,
		c.Version,
		c.UserIP,
	)

	if err != nil {
		logger.WithField("trace", "SaveAccess").Error(err)
		return err
	}

	return nil
}

// LoadAccess retrieves previously saved access token by open id.
func (env Env) LoadAccess(openID string, c util.RequestClient) (OAuthAccess, error) {
	query := `
	SELECT access_token AS accessToken,
		expires_in AS expiresIn,
		refresh_token AS refreshToken,
		open_id AS opendId,
		scope AS scope,
		union_id AS unionId
	FROM user_db.wechat_access
	WHERE open_id = ?
		AND client_type = ?
	ORDER BY updated_utc DESC
	LIMIT 1`

	var acc OAuthAccess
	err := env.DB.QueryRow(query, openID, c.ClientType).Scan(
		&acc.AccessToken,
		&acc.ExpiresIn,
		&acc.RefreshToken,
		&acc.OpenID,
		&acc.Scope,
		&acc.UnionID,
	)

	if err != nil {
		logger.WithField("trace", "LoadAccess").Error(err)
		return acc, err
	}

	return acc, nil
}

// UpdateAccess saves refreshed access token.
func (env Env) UpdateAccess(openID, acc OAuthAccess, c util.RequestClient) error {
	query := `
	UPDATE user_db.wechat_access
	SET access_token = ?,
		expires_in = ?,
		open_id = ?,
		scope = ?,
		client_version = ?,
		user_ip = INET6_ATON(?)
	WHERE open_id = ?
	LIMIT 1`

	_, err := env.DB.Exec(query,
		acc.AccessToken,
		acc.ExpiresIn,
		acc.OpenID,
		acc.Scope,
		c.Version,
		c.UserIP,
		openID,
	)

	if err != nil {
		logger.WithField("trace", "UpdateAccess").Error(err)
		return err
	}

	return nil
}
