package wxlogin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/spf13/viper"
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

var keys = []string{
	// 移动应用 -> FT中文网会员订阅. This is used for Android subscription
	"wxapp.native_app",
	// 移动应用 -> FT中文网. This is for iOS subscription and legacy Android subscription.
	"wxapp.web_pay",
	// 网站应用 -> FT中文网. This is used for web login
	"wxapp.web_oauth",
}

func MustInitApps() map[string]OAuthApp {
	apps := make(map[string]OAuthApp)

	for _, k := range keys {
		app := MustNewOAuthApp(k)

		apps[app.AppID] = app
	}

	return apps
}

// OAuthApp contains essential credentials to call Wecaht API.
type OAuthApp struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"secret"`
}

func NewOAuthApp(key string) (OAuthApp, error) {
	var app OAuthApp
	err := viper.UnmarshalKey(key, &app)
	if err != nil {
		return OAuthApp{}, err
	}

	if err := app.Validate(); err != nil {
		return OAuthApp{}, err
	}

	return app, nil
}

func MustNewOAuthApp(key string) OAuthApp {
	app, err := NewOAuthApp(key)
	if err != nil {
		panic(err)
	}

	return app
}

// Validate makes sure AppID and AppSecret is properly configured.
func (a OAuthApp) Validate() error {
	if a.AppID == "" || a.AppSecret == "" {
		return errors.New("wechat oauth app id or secret cannot be empty")
	}

	return nil
}

// Build url to get access token
func (a OAuthApp) accessTokenURL(code string) string {
	q := url.Values{}
	q.Set("appid", a.AppID)
	q.Set("secret", a.AppSecret)
	q.Set("code", code)
	q.Set("grant_type", "authorization_code")

	return fmt.Sprintf("%s/oauth2/access_token?%s", apiBaseURL, q.Encode())
}

func (a OAuthApp) userInfoURL(accessToken, openID string) string {
	q := url.Values{}
	q.Set("access_token", accessToken)
	q.Set("openid", openID)

	return fmt.Sprintf("%s/userinfo?%s", apiBaseURL, q.Encode())
}

func (a OAuthApp) refreshTokeURL(token string) string {
	q := url.Values{}
	q.Set("appid", a.AppID)
	q.Set("grant_type", "refresh_token")
	q.Set("refresh_token", token)

	return fmt.Sprintf("%s/oauth2/refresh_token?%s", apiBaseURL, q.Encode())
}

func (a OAuthApp) accessValidityURL(accessToken, openID string) string {
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
func (a OAuthApp) GetAccessToken(code string) (OAuthAccess, error) {
	u := a.accessTokenURL(code)

	var acc OAuthAccess
	_, body, errs := request.Get(u).Set("Accept-Language", "en-US,en;q=0.5").End()

	if errs != nil {
		logger.WithField("trace", "OAuthApp.GetAccessToken").Error(errs)

		return acc, errs[0]
	}

	// {"access_token":"***REMOVED***",
	// "expires_in":7200,
	// "refresh_token":"22_FfPqWuDBKDZtCwsTyO9tCtWolvi62kXTioDSKN-OO00xxQcLCovxWxg_FWt17Ca5chDjKiQ_aQMyErN4NIJYTCMI0VAcN2Z5Yv2W9kj-AyM",
	// "openid":"ofP-k1LSVS-ObmrySM1aXKbv1Hjs",
	// "scope":"snsapi_login",
	// "unionid":"ogfvwjk6bFqv2yQpOrac0J3PqA0o"}
	logger.WithField("trace", "OAuthApp.GetAccessToken").Infof("Wechat response: %s\n", body)

	if err := json.Unmarshal([]byte(body), &acc); err != nil {
		logger.WithField("trace", "OAuthApp.GetAccessToken").Error(errs)
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
func (a OAuthApp) GetUserInfo(accessToken, openID string) (UserInfo, error) {
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
	logger.WithField("trace", "OAuthApp.GetUserInfo").Infof("Wechat user info: %s", body)

	if errs != nil {
		logger.WithField("trace", "OAuthApp.GetUserInfo").Error(errs)

		return info, errs[0]
	}

	if err := json.Unmarshal([]byte(body), &info); err != nil {
		logger.WithField("trace", "OAuthApp.GetUserInfo").Error(errs)
		return info, err
	}

	return info, nil
}

// RefreshAccess refresh access token.
func (a OAuthApp) RefreshAccess(refreshToken string) (OAuthAccess, error) {
	u := a.refreshTokeURL(refreshToken)

	var acc OAuthAccess
	_, body, errs := request.Get(u).Set("Accept-Language", "en-US,en;q=0.5").End()

	logger.WithField("trace", "OAuthApp.RefreshAccess").Infof("Response: %s", body)

	if errs != nil {
		logger.WithField("trace", "OAuthApp.RefreshAccess").Error(errs)

		return acc, errs[0]
	}

	if err := json.Unmarshal([]byte(body), &acc); err != nil {
		logger.WithField("trace", "OAuthApp.RefreshAccess").Error(err)

		return acc, err
	}

	return acc, nil
}

// IsValidAccess checks if an access token is valid.
func (a OAuthApp) IsValidAccess(accessToken, openID string) bool {
	u := a.accessValidityURL(accessToken, openID)

	var resp RespStatus

	_, body, errs := request.Get(u).Set("Accept-Language", "en-US,en;q=0.5").End()

	if errs != nil {
		logger.WithField("trace", "OAuthApp.IsInvalidAccess").Error(errs)
		return false
	}

	logger.WithField("trace", "OAuthApp.IsValidAccess").Infof("Response: %s", body)

	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		logger.WithField("trace", "OAuthApp.IsValidAccess").Error(err)
		return false
	}

	if resp.Code == 0 && resp.Message == "ok" {
		return true
	}

	return false
}
