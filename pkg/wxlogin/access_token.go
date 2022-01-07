package wxlogin

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/guregu/null"
	"time"
)

// AccessResponse is the response from Wechat API.
// Return from
// /sns/oauth2/access_token?appid=APPID&secret=SECRET&code=CODE&grant_type=authorization_code
// and
// /sns/oauth2/refresh_token?appid=APPID&grant_type=refresh_token&refresh_token=REFRESH_TOKEN
type AccessResponse struct {
	// Example: 16_Ix0E3WfWs9u5Rh9f-lB7_LgsQJ4zm1eodolFJpSzoQibTAuhIlp682vDmkZSaYIjD9gekOa1zQl-6c6S_CrN_cN9vx9mybwXNVgFbwPMMwM
	AccessToken string `json:"access_token" db:"access_token"`
	// Example: 7200
	ExpiresIn int64 `json:"expires_in" db:"expires_in"`
	// Exmaple: 16_IlmA9eLGjJw7gBKBT48wff1V1hAYAdpmIqUAypspepm6DsQ6kkcLeZmP932s9PcKp1WM5P_1YwUNQqF-29B_0CqGTqMpWkaaiNSYp26MmB4
	RefreshToken string `json:"refresh_token" db:"refresh_token"`
	// Example: ob7fA0h69OO0sTLyQQpYc55iF_P0
	OpenID string `json:"openid" db:"open_id"`
	// Example: snsapi_userinfo
	Scope string `json:"scope" db:"scope"`
	// Example: String:ogfvwjk6bFqv2yQpOrac0J3PqA0o Valid:true
	UnionID null.String `json:"unionid" db:"union_id"`
	RespStatus
}

func (a AccessResponse) UserInfoParams() UserInfoParams {
	return UserInfoParams{
		AccessToken: a.AccessToken,
		OpenID:      a.OpenID,
	}
}

// GenerateSessionID generate an id used to identify this unique session.
// Concatenate AccessToken, RefreshToken, OpenID by a colon and generate the MD5 sum of the string.
// Returns the hexadecimal encoded string of the MD5 bytes.
// Database should use VARBINARY(16) to store this value.
// This id is only generated once during the lifetime
// of a refresh token.
// When access token is refreshed, this id won't be changed.
func (a AccessResponse) GenerateSessionID() string {
	data := fmt.Sprintf("%s:%s:%s", a.AccessToken, a.RefreshToken, a.OpenID)
	h := md5.Sum([]byte(data))
	return hex.EncodeToString(h[:])
}

type AccessSchema struct {
	ID    string `db:"session_id"`
	AppID string `db:"app_id"`
	AccessResponse
	footprint.Client
	CreatedUTC chrono.Time `db:"created_utc"`
	UpdatedUTC chrono.Time `db:"updated_utc"`
}

func NewAccessSchema(r AccessResponse, appID string, client footprint.Client) AccessSchema {
	return AccessSchema{
		ID:             r.GenerateSessionID(),
		AppID:          appID,
		AccessResponse: r,
		Client:         client,
		CreatedUTC:     chrono.TimeNow(),
		UpdatedUTC:     chrono.TimeNow(),
	}
}

// WithAccessToken refreshes the access token.
func (s AccessSchema) WithAccessToken(t string) AccessSchema {
	s.AccessToken = t
	s.UpdatedUTC = chrono.TimeNow()
	return s
}

// IsAccessExpired tests if access token is expired.
func (s AccessSchema) IsAccessExpired() bool {
	after2Hours := s.UpdatedUTC.Add(time.Second * time.Duration(s.ExpiresIn))
	return after2Hours.Before(time.Now())
}

// IsRefreshExpired tests if refresh token is expired.
func (s AccessSchema) IsRefreshExpired() bool {
	after30Days := s.CreatedUTC.AddDate(0, 0, 30)
	return after30Days.Before(time.Now())
}

// Session is a reduced version of AccessSchema sent to client.
// A Wechat login session should expires in 30 days,
// which is the duration of refresh token.
type Session struct {
	ID        string      `json:"sessionId"` // use this to refresh account.
	UnionID   string      `json:"unionId"`   // Use this to fetch account data
	CreatedAt chrono.Time `json:"createdAt"`
}

func NewSession(s AccessSchema, unionID string) Session {
	return Session{
		ID:        s.ID,
		UnionID:   unionID,
		CreatedAt: s.CreatedUTC,
	}
}
