package wxoauth

import (
	"github.com/FTChinese/subscription-api/pkg/client"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
)

// SaveWxAccess saves the access token related data after acquired from wechat api.
func (env Env) SaveWxAccess(appID string, acc wxlogin.OAuthAccess, c client.Client) error {

	_, err := env.dbs.Write.Exec(wxlogin.StmtInsertAccess,
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
		return err
	}

	return nil
}

// LoadWxAccess retrieves previously saved access token by open id.
// Is it possible wechat generate different openID under the same app?
// Or is it possible wechat generate same openID for different app?
// What if iOS and Android used the same Wechat app? Will they use the same access token or different one?
// Open ID is always the same the under the same Wechat app.
func (env Env) LoadWxAccess(appID, sessionID string) (wxlogin.OAuthAccess, error) {

	var acc wxlogin.OAuthAccess
	err := env.dbs.Read.QueryRow(wxlogin.StmtSelectAccess,
		sessionID,
		appID,
	).Scan(
		&acc.SessionID,
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
		return acc, err
	}

	return acc, nil
}

// UpdateWxAccess saves refreshed access token.
func (env Env) UpdateWxAccess(sessionID, accessToken string) error {

	_, err := env.dbs.Write.Exec(wxlogin.StmtUpdateAccess,
		accessToken,
		sessionID,
	)

	if err != nil {
		return err
	}

	return nil
}
