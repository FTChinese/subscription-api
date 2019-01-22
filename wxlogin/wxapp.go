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
	appID     string
	appSecret string
}

// NewWxApp creates a new Wecaht client.
func NewWxApp(id, secret string) (WxApp, error) {
	if id == "" || secret == "" {
		return WxApp{}, errors.New("app id and secret must not be empty")
	}
	return WxApp{
		appID:     id,
		appSecret: secret,
	}, nil
}

// Build url to get access token
func (c WxApp) accessTokenURL(code string) string {
	q := url.Values{}
	q.Set("appid", c.appID)
	q.Set("secret", c.appSecret)
	q.Set("code", code)
	q.Set("grant_type", "authorization_code")

	return fmt.Sprintf("%s/oauth2/access_token?%s", apiBaseURL, q.Encode())
}

func (c WxApp) userInfoURL(accessToken, openID string) string {
	q := url.Values{}
	q.Set("access_token", accessToken)
	q.Set("openid", openID)

	return fmt.Sprintf("%s/userinfo?%s", apiBaseURL, q.Encode())
}

func (c WxApp) refreshTokeURL(token string) string {
	q := url.Values{}
	q.Set("appid", c.appID)
	q.Set("grant_type", "refresh_token")
	q.Set("refresh_token", token)

	return fmt.Sprintf("%s/oauth2/refresh_token?%s", apiBaseURL, q.Encode())
}

func (c WxApp) accessValidityURL(accessToken, openID string) string {
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
func (c WxApp) GetAccessToken(code string) (OAuthAccess, error) {
	u := c.accessTokenURL(code)

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
func (c WxApp) GetUserInfo(accessToken, openID string) (UserInfo, error) {
	u := c.userInfoURL(accessToken, openID)

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
func (c WxApp) RefreshAccess(refreshToken string) (OAuthAccess, error) {
	u := c.refreshTokeURL(refreshToken)

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
func (c WxApp) IsValidAccess(accessToken, openID string) bool {
	u := c.accessValidityURL(accessToken, openID)

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
