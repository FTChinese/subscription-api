package wxlogin

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/util"
)

// Session records a Wechat login session.
// A Wechat login session should expires in 30 days,
// which is the duration of refresh token.
type Session struct {
	ID        string    `json:"id"`
	UnionID   string    `json:"unionId"`
	CreatedAt util.Time `json:"createdAt"`
}

// OAuthAccess is the response of Wechat endpoint
// /sns/oauth2/access_token?appid=APPID&secret=SECRET&code=CODE&grant_type=authorization_code
// and
// /sns/oauth2/refresh_token?appid=APPID&grant_type=refresh_token&refresh_token=REFRESH_TOKEN
type OAuthAccess struct {
	SessionID string
	// Example: ***REMOVED***
	AccessToken string `json:"access_token"`
	// Example: 7200
	ExpiresIn int64 `json:"expires_in"`
	// Exmaple: ***REMOVED***
	RefreshToken string `json:"refresh_token"`
	// Example: ob7fA0h69OO0sTLyQQpYc55iF_P0
	OpenID string `json:"openid"`
	// Example: snsapi_userinfo
	Scope string `json:"scope"`
	// Example: String:ogfvwjk6bFqv2yQpOrac0J3PqA0o Valid:true
	UnionID   null.String `json:"unionid"`
	createdAt util.Time
	updatedAt util.Time
	RespStatus
}

// ToSession creates a Session instance.
func (a *OAuthAccess) ToSession(unionID string) Session {
	return Session{
		ID:        a.SessionID,
		UnionID:   unionID,
		CreatedAt: a.createdAt,
	}
}

// generateSessionID generate an id used to identify this unique session.
// Concatenate AccessToken, RefreshToken, OpenID by a colon and generate the MD5 sum of the string.
// Returns the hexadecmial encoded string of the MD5 bytes.
// Database should use VARBINARY(16) to store this value.
func (a *OAuthAccess) generateSessionID() {
	data := fmt.Sprintf("%s:%s:%s", a.AccessToken, a.RefreshToken, a.OpenID)
	h := md5.Sum([]byte(data))
	a.SessionID = fmt.Sprintf("%x", h)
}

// IsAccessExpired tests if access token is expired.
func (a OAuthAccess) IsAccessExpired() bool {
	after2Hours := a.updatedAt.Add(time.Second * time.Duration(a.ExpiresIn))
	return after2Hours.Before(time.Now())
}

// IsRefreshExpired tests if refresh token is expired.
func (a OAuthAccess) IsRefreshExpired() bool {
	after30Days := a.createdAt.AddDate(0, 0, 30)
	return after30Days.Before(time.Now())
}

// SaveAccess saves the access token related data after acquired from wechat api.
// Or update access token by refresh token if it is expired.
// For every authoriztion request, a new pair of access token and refresh token are generated, even on the same platform under single wechat app.
// Returns a session id that uniquely identify this row.
func (env Env) SaveAccess(appID string, acc OAuthAccess, c util.ClientApp) error {
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

	_, err := env.DB.Exec(query,
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
		acc.createdAt,
		acc.updatedAt,
	)

	if err != nil {
		logger.WithField("trace", "SaveAccess").Error(err)
		return err
	}

	return nil
}

// LoadAccess retrieves previously saved access token by open id.
// Is it possbile wechat generate different openID under the same app?
// Or is it possible wechat generate same openID for different app?
// What if iOS and Android used the same Wechat app? Will they use the same access token or different one?
// Open ID is always the same for the under the same Wechat app.
func (env Env) LoadAccess(appID, sessionID string) (OAuthAccess, error) {
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

	var acc OAuthAccess
	err := env.DB.QueryRow(query, sessionID, appID).Scan(
		&acc.AccessToken,
		&acc.ExpiresIn,
		&acc.RefreshToken,
		&acc.OpenID,
		&acc.Scope,
		&acc.UnionID,
		&acc.createdAt,
		&acc.updatedAt,
	)

	if err != nil {
		logger.WithField("trace", "LoadAccess").Error(err)
		return acc, err
	}

	return acc, nil
}

// UpdateAccess saves refreshed access token.
func (env Env) UpdateAccess(sessionID, accessToken string) error {
	query := `
	UPDATE user_db.wechat_access
	SET access_token = ?,
	WHERE session_id = UNHEX(?)
	LIMIT 1`

	_, err := env.DB.Exec(query,
		accessToken,
		sessionID,
	)

	if err != nil {
		logger.WithField("trace", "UpdateAccess").Error(err)
		return err
	}

	return nil
}
