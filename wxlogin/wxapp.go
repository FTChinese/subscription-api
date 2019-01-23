package wxlogin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"net/url"

	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
)

var (
	logger  = logrus.WithField("package", "wxlogin")
	request = gorequest.New()
)

const (
	apiBaseURL = "https://api.weixin.qq.com/sns"
)

// WxApp contains essential credentials to call Wecaht API.
type WxApp struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"secret"`
}

// Ensure makes sure AppID and AppSecret is properly configured.
func (a WxApp) Ensure() error {
	if a.AppID == "" || a.AppSecret == "" {
		return errors.New("wechat oauth app id or secret cannot be empty")
	}

	return nil
}

// Build url to get access token
func (a WxApp) accessTokenURL(code string) string {
	q := url.Values{}
	q.Set("appid", a.AppID)
	q.Set("secret", a.AppSecret)
	q.Set("code", code)
	q.Set("grant_type", "authorization_code")

	return fmt.Sprintf("%s/oauth2/access_token?%s", apiBaseURL, q.Encode())
}

func (a WxApp) userInfoURL(accessToken, openID string) string {
	q := url.Values{}
	q.Set("access_token", accessToken)
	q.Set("openid", openID)

	return fmt.Sprintf("%s/userinfo?%s", apiBaseURL, q.Encode())
}

func (a WxApp) refreshTokeURL(token string) string {
	q := url.Values{}
	q.Set("appid", a.AppID)
	q.Set("grant_type", "refresh_token")
	q.Set("refresh_token", token)

	return fmt.Sprintf("%s/oauth2/refresh_token?%s", apiBaseURL, q.Encode())
}

func (a WxApp) accessValidityURL(accessToken, openID string) string {
	q := url.Values{}
	q.Set("access_token", accessToken)
	q.Set("openid", openID)

	return fmt.Sprintf("%s/auth?%s", apiBaseURL, q.Encode())
}

// GetAccessToken request for access token with a code previsouly acquired from wechat.
// For every authoriztion request, a new pair of access token and refresh token are generated, even on the same platform under single wechat app.
//
// Possible error response:
// errcode: 41002, errmsg: "appid missing";
// errcode: 40029, errmsg: "invalid code";
// Response without error: errcode: 0, errmsg: "";
// What will be returned if two different code under the same Wechat app applied for access token simutaneously?
func (a WxApp) GetAccessToken(code string) (OAuthAccess, error) {
	u := a.accessTokenURL(code)

	var acc OAuthAccess
	_, body, errs := request.Get(u).End()

	if errs != nil {
		logger.WithField("trace", "GetAccessToken").Error(errs)

		return acc, errs[0]
	}

	logger.Infof("Response body: %s\n", body)

	if err := json.Unmarshal([]byte(body), &acc); err != nil {
		logger.WithField("trace", "GetAccessToken").Error(errs)
		return acc, err
	}

	acc.GenerateSessionID()
	acc.CreatedAt = chrono.TimeNow()
	acc.UpdatedAt = chrono.TimeNow()

	return acc, nil
}

// GetUserInfo from Wechat by open id.
func (a WxApp) GetUserInfo(accessToken, openID string) (UserInfo, error) {
	u := a.userInfoURL(accessToken, openID)

	var info UserInfo

	_, body, errs := request.Get(u).End()

	if errs != nil {
		logger.WithField("trace", "GetUserInfo").Error(errs)

		return info, errs[0]
	}

	if err := json.Unmarshal([]byte(body), &info); err != nil {
		logger.WithField("trace", "GetUserInfo").Error(errs)
		return info, err
	}

	return info, nil
}

// RefreshAccess refresh access token.
func (a WxApp) RefreshAccess(refreshToken string) (OAuthAccess, error) {
	u := a.refreshTokeURL(refreshToken)

	var acc OAuthAccess
	_, body, errs := request.Get(u).End()

	if errs != nil {
		logger.WithField("trace", "RefreshAccess").Error(errs)

		return acc, errs[0]
	}

	if err := json.Unmarshal([]byte(body), &acc); err != nil {
		logger.WithField("trace", "RefreshAccess").Error(err)

		return acc, err
	}

	return acc, nil
}

// IsValidAccess checks if an access token is valid.
func (a WxApp) IsValidAccess(accessToken, openID string) bool {
	u := a.accessValidityURL(accessToken, openID)

	var resp RespStatus

	_, body, errs := request.Get(u).End()

	if errs != nil {
		logger.WithField("trace", "IsInvalidAccess").Error(errs)
		return false
	}

	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		logger.WithField("trace", "IsValidAccess")
		return false
	}

	if resp.Code == 0 && resp.Message == "ok" {
		return true
	}

	return false
}
