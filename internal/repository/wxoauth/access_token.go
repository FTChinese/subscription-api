package wxoauth

import (
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
)

// SaveWxAccess saves the access token related data after acquired from wechat api.
func (env Env) SaveWxAccess(s wxlogin.AccessSchema) error {

	_, err := env.dbs.Write.NamedExec(
		wxlogin.StmtInsertAccess,
		s,
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
func (env Env) LoadWxAccess(appID, sessionID string) (wxlogin.AccessSchema, error) {

	var acc wxlogin.AccessSchema
	err := env.dbs.Read.Get(
		&acc,
		wxlogin.StmtSelectAccess,
		sessionID,
		appID,
	)

	if err != nil {
		return acc, err
	}

	return acc, nil
}

// UpdateWxAccess saves refreshed access token.
func (env Env) UpdateWxAccess(s wxlogin.AccessSchema) error {

	_, err := env.dbs.Write.NamedExec(
		wxlogin.StmtUpdateAccess,
		s,
	)

	if err != nil {
		return err
	}

	return nil
}
