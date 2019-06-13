package model

import (
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/wxlogin"
)

// SaveWxAccess saves the access token related data after acquired from wechat api.
func (env Env) SaveWxAccess(appID string, acc wxlogin.OAuthAccess, c util.ClientApp) error {

	_, err := env.db.Exec(query.InsertWxAccess,
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
		logger.WithField("trace", "Env.SaveAccess").Error(err)
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
	err := env.db.QueryRow(query.SelectWxAccess,
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
		logger.WithField("trace", "Env.LoadAccess").Error(err)
		return acc, err
	}

	return acc, nil
}

// UpdateWxAccess saves refreshed access token.
func (env Env) UpdateWxAccess(sessionID, accessToken string) error {

	_, err := env.db.Exec(query.UpdateWxAccess,
		accessToken,
		sessionID,
	)

	if err != nil {
		logger.WithField("trace", "Env.UpdateAccess").Error(err)
		return err
	}

	return nil
}

// SaveWxUser from wechat API.
// Since a user can authorize multiple times, use ON DUPLICATE to handle unique key constraint.
func (env Env) SaveWxUser(u wxlogin.UserInfo) error {

	_, err := env.db.Exec(query.InsertWxUser,
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
		logger.WithField("trace", "Env.SaveWxUser").Error(err)
		return err
	}

	return nil
}

// UpdateWxUser update data of one union id.
func (env Env) UpdateWxUser(u wxlogin.UserInfo) error {

	_, err := env.db.Exec(query.UpdateWxUser,
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
		logger.WithField("trace", "Env.UpdateUserInfo").Error(err)
		return err
	}

	return nil
}

// SaveWxStatus saves Wechat OAuth API error response into data so that we could know what errors on hell Wechat actually produced.
func (env Env) SaveWxStatus(code int64, message string) error {

	_, err := env.db.Exec(query.InsertWxStatus,
		code,
		message,
	)

	if err != nil {
		logger.WithField("trace", "Env.SaveWxError").Error(err)
		return err
	}

	return nil
}
