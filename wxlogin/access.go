package wxlogin

import "gitlab.com/ftchinese/subscription-api/util"

// OAuthAccess contains data returned by exchange access token with oauth code.
type OAuthAccess struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
}

// SaveAccess saves the access token related data after acquired from wechat api.
func (env Env) SaveAccess(acc OAuthAccess, c util.RequestClient) error {
	query := `INSERT INTO user_db.wechat_access
	SET access_token = ?,
		expires_in = ?,
		refresh_token = ?,
		open_id = ?,
		scope = ?,
		union_id = NULLIF(?, ''),
		client_type = NULLIF(?, ''),
		client_version = NULLIF(?, ''),
		user_ip = NULLIF(INET6_ATON(?), NULL)
	ON DUPLICATE KEY UPDATE access_token = ?,
		expires_in = ?,
		refresh_token = ?,
		client_type = NULLIF(?, ''),
		client_version = NULLIF(?, ''),
		user_ip = NULLIF(INET6_ATON(?), NULL)`

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
