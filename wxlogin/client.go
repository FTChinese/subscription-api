package wxlogin

import (
	"encoding/json"
	"fmt"
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

// RespStatus is used to parse wechat error response.
type RespStatus struct {
	Code    int64  `json:"errcode"`
	Message string `json:"errmsg"`
}

// IsError tests if wechat api response is an error.
// Wechat does not tell what error will returned exactly.
// It does not use restful standards. You cannot rely on HTTP status codes.
func (r RespStatus) IsError() bool {
	return !(r.Code == 0 && (r.Message == "ok" || r.Message == ""))
}

// Client contains essential credentials to call Wecaht API.
type Client struct {
	AppID     string
	AppSecret string
}

// NewClient creates a new Wecaht client.
func NewClient(id, secret string) Client {
	return Client{
		AppID:     id,
		AppSecret: secret,
	}
}

// Build url to get access token
func (c Client) accessTokenURL(code string) string {
	q := url.Values{}
	q.Set("appid", c.AppID)
	q.Set("secret", c.AppSecret)
	q.Set("code", code)
	q.Set("grant_type", "authorization_code")

	return fmt.Sprintf("%s/oauth2/access_token?%s", apiBaseURL, q.Encode())
}

func (c Client) refreshTokeURL(token string) string {
	q := url.Values{}
	q.Set("appid", c.AppID)
	q.Set("grant_type", "refresh_token")
	q.Set("refresh_token", token)

	return fmt.Sprintf("%s/oauth2/refresh_token?%s", apiBaseURL, q.Encode())
}

func (c Client) userInfoURL(token, openID string) string {
	q := url.Values{}
	q.Set("access_token", token)
	q.Set("openid", openID)

	return fmt.Sprintf("%s/userinfo?%s", apiBaseURL, q.Encode())
}

func (c Client) accessValidityURL(accessToken, openID string) string {
	q := url.Values{}
	q.Set("access_token", accessToken)
	q.Set("openid", openID)

	return fmt.Sprintf("%s/auth?%s", apiBaseURL, q.Encode())
}

// GetAccessToken request for access token with a code previsouly acquired from wechat.
// Possible error response:
// errcode: 41002, errmsg: appid missing;
// errcode: 40029, errmsg: invalid code;
// Response without error: errcode: 0, errmsg: "";
func (c Client) GetAccessToken(code string) (OAuthAccess, error) {
	u := c.accessTokenURL(code)

	var acc OAuthAccess
	_, body, errs := request.Get(u).End()

	if errs != nil {
		logger.WithField("trace", "GetAccessToken").Error(errs)

		return acc, errs[0]
	}

	if err := json.Unmarshal([]byte(body), &acc); err != nil {
		logger.WithField("trace", "GetAccessToken").Error(errs)
		return acc, err
	}

	return acc, nil
}

// GetUserInfo from Wechat by open id.
func (c Client) GetUserInfo(acc OAuthAccess) (UserInfo, error) {
	// First use openId to retrieve access token

	// Then build url
	u := c.userInfoURL(acc.AccessToken, acc.OpenID)

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
func (c Client) RefreshAccess(refreshToken string) (OAuthAccess, error) {
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
func (c Client) IsValidAccess(acc OAuthAccess) bool {
	u := c.accessValidityURL(acc.AccessToken, acc.OpenID)

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
