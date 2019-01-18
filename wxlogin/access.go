package wxlogin

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/guregu/null"
)

// Session records a Wechat login session.
// A Wechat login session should expires in 30 days,
// which is the duration of refresh token.
type Session struct {
	ID        string      `json:"sessionId"`
	UnionID   string      `json:"unionId"`
	CreatedAt chrono.Time `json:"createdAt"`
}

// OAuthAccess is the response of Wechat endpoint
// /sns/oauth2/access_token?appid=APPID&secret=SECRET&code=CODE&grant_type=authorization_code
// and
// /sns/oauth2/refresh_token?appid=APPID&grant_type=refresh_token&refresh_token=REFRESH_TOKEN
type OAuthAccess struct {
	SessionID string
	// Example: 16_Ix0E3WfWs9u5Rh9f-lB7_LgsQJ4zm1eodolFJpSzoQibTAuhIlp682vDmkZSaYIjD9gekOa1zQl-6c6S_CrN_cN9vx9mybwXNVgFbwPMMwM
	AccessToken string `json:"access_token"`
	// Example: 7200
	ExpiresIn int64 `json:"expires_in"`
	// Exmaple: 16_IlmA9eLGjJw7gBKBT48wff1V1hAYAdpmIqUAypspepm6DsQ6kkcLeZmP932s9PcKp1WM5P_1YwUNQqF-29B_0CqGTqMpWkaaiNSYp26MmB4
	RefreshToken string `json:"refresh_token"`
	// Example: ob7fA0h69OO0sTLyQQpYc55iF_P0
	OpenID string `json:"openid"`
	// Example: snsapi_userinfo
	Scope string `json:"scope"`
	// Example: String:ogfvwjk6bFqv2yQpOrac0J3PqA0o Valid:true
	UnionID   null.String `json:"unionid"`
	CreatedAt chrono.Time
	UpdatedAt chrono.Time
	RespStatus
}

// ToSession creates a Session instance.
func (a *OAuthAccess) ToSession(unionID string) Session {
	return Session{
		ID:        a.SessionID,
		UnionID:   unionID,
		CreatedAt: a.CreatedAt,
	}
}

// GenerateSessionID generate an id used to identify this unique session.
// Concatenate AccessToken, RefreshToken, OpenID by a colon and generate the MD5 sum of the string.
// Returns the hexadecmial encoded string of the MD5 bytes.
// Database should use VARBINARY(16) to store this value.
func (a *OAuthAccess) GenerateSessionID() {
	if a.SessionID != "" {
		return
	}
	data := fmt.Sprintf("%s:%s:%s", a.AccessToken, a.RefreshToken, a.OpenID)
	h := md5.Sum([]byte(data))
	a.SessionID = hex.EncodeToString(h[:])
}

// IsAccessExpired tests if access token is expired.
func (a OAuthAccess) IsAccessExpired() bool {
	after2Hours := a.UpdatedAt.Add(time.Second * time.Duration(a.ExpiresIn))
	return after2Hours.Before(time.Now())
}

// IsRefreshExpired tests if refresh token is expired.
func (a OAuthAccess) IsRefreshExpired() bool {
	after30Days := a.CreatedAt.AddDate(0, 0, 30)
	return after30Days.Before(time.Now())
}
