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

// Client contains essential credentials to call Wecaht API.
type Client struct {
	AppID     string
	AppSecret string
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

// GetAccessToken request for access token with a code previsouly acquired from wechat.
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
