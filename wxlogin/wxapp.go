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

// GetAccessToken request for access token with a code previously acquired from wechat.
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
	_, body, errs := request.Get(u).Set("Accept-Language", "en-US,en;q=0.5").End()

	if errs != nil {
		logger.WithField("trace", "WxApp.GetAccessToken").Error(errs)

		return acc, errs[0]
	}

	// {"access_token":"22_JJVz_GH32Bt89Cfj_kaSYr5V-j8_iNphiYzQ3i3rMNRdk88k8GZw_v5qhuR9e3X5mZtn4-QTIyqgzmruxSlVZ0shrU9v3mzV7dLY46t4K0M",
	// "expires_in":7200,
	// "refresh_token":"22_FfPqWuDBKDZtCwsTyO9tCtWolvi62kXTioDSKN-OO00xxQcLCovxWxg_FWt17Ca5chDjKiQ_aQMyErN4NIJYTCMI0VAcN2Z5Yv2W9kj-AyM",
	// "openid":"ofP-k1LSVS-ObmrySM1aXKbv1Hjs",
	// "scope":"snsapi_login",
	// "unionid":"ogfvwjk6bFqv2yQpOrac0J3PqA0o"}
	logger.WithField("trace", "WxApp.GetAccessToken").Infof("Wechat response: %s\n", body)

	if err := json.Unmarshal([]byte(body), &acc); err != nil {
		logger.WithField("trace", "WxApp.GetAccessToken").Error(errs)
		return acc, err
	}

	// Create an session id to identify this unique session.
	acc.GenerateSessionID()
	acc.CreatedAt = chrono.TimeNow()
	acc.UpdatedAt = chrono.TimeNow()

	return acc, nil
}

// GetUserInfo from Wechat by open id.
// It seems wechat return empty fields as empty string.
func (a WxApp) GetUserInfo(accessToken, openID string) (UserInfo, error) {
	u := a.userInfoURL(accessToken, openID)

	var info UserInfo

	_, body, errs := request.Get(u).Set("Accept-Language", "en-US,en;q=0.5").End()

	// {
	// "openid":"ofP-k1LSVS-ObmrySM1aXKbv1Hjs",
	// "nickname":"倪卫国的小号",
	// "sex":0,
	// "language":"zh_CN",
	// "city":"",
	// "province":"",
	// "country":"",
	// "headimgurl":"http://thirdwx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTLvOQseDrWKNS8H4msGicM2EI4DdaC5q5dzSoV8icVicmde2rTaERvXGG7jbLOk89Ish5ppRy1rVGIDA\\/132",
	// "privilege":[],
	// "unionid":"ogfvwjk6bFqv2yQpOrac0J3PqA0o"
	// }
	logger.WithField("trace", "WxApp.GetUserInfo").Infof("Wechat user info: %s", body)

	if errs != nil {
		logger.WithField("trace", "WxApp.GetUserInfo").Error(errs)

		return info, errs[0]
	}

	if err := json.Unmarshal([]byte(body), &info); err != nil {
		logger.WithField("trace", "WxApp.GetUserInfo").Error(errs)
		return info, err
	}

	return info, nil
}

// RefreshAccess refresh access token.
func (a WxApp) RefreshAccess(refreshToken string) (OAuthAccess, error) {
	u := a.refreshTokeURL(refreshToken)

	var acc OAuthAccess
	_, body, errs := request.Get(u).Set("Accept-Language", "en-US,en;q=0.5").End()

	logger.WithField("trace", "WxApp.RefreshAccess").Infof("Response: %s", body)

	if errs != nil {
		logger.WithField("trace", "WxApp.RefreshAccess").Error(errs)

		return acc, errs[0]
	}

	if err := json.Unmarshal([]byte(body), &acc); err != nil {
		logger.WithField("trace", "WxApp.RefreshAccess").Error(err)

		return acc, err
	}

	return acc, nil
}

// IsValidAccess checks if an access token is valid.
func (a WxApp) IsValidAccess(accessToken, openID string) bool {
	u := a.accessValidityURL(accessToken, openID)

	var resp RespStatus

	_, body, errs := request.Get(u).Set("Accept-Language", "en-US,en;q=0.5").End()

	if errs != nil {
		logger.WithField("trace", "WxApp.IsInvalidAccess").Error(errs)
		return false
	}

	logger.WithField("trace", "WxApp.IsValidAccess").Infof("Response: %s", body)

	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		logger.WithField("trace", "WxApp.IsValidAccess").Error(err)
		return false
	}

	if resp.Code == 0 && resp.Message == "ok" {
		return true
	}

	return false
}
