package wxoauth

import (
	"encoding/json"
	"github.com/FTChinese/subscription-api/lib/fetch"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"go.uber.org/zap"
)

const (
	apiBaseURL = "https://api.weixin.qq.com/sns"
	acceptLang = "en-US,en;q=0.5"
)

// Client is used to access wechat API.
type Client struct {
	app    config.WechatApp
	logger *zap.Logger
}

func NewClient(app config.WechatApp, logger *zap.Logger) Client {
	return Client{
		app:    app,
		logger: logger,
	}
}

func (c Client) GetApp() config.WechatApp {
	return c.app
}

// GetAccessToken request for access token with a code previously acquired from wechat.
// For every authoriztion request, a new pair of access token and refresh token are generated, even on the same platform under single wechat app.
//
// Possible error response:
// errcode: 41002, errmsg: "appid missing";
// errcode: 40029, errmsg: "invalid code";
// Response without error: errcode: 0, errmsg: "";
// What will be returned if two different code under the same Wechat app applied for access token simutaneously?
func (c Client) GetAccessToken(code string) (wxlogin.AccessResponse, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	resp, errs := fetch.New().
		Get(apiBaseURL + "/oauth2/access_token").
		SetQueryN(map[string]string{
			"appid":      c.app.AppID,
			"secret":     c.app.AppSecret,
			"code":       code,
			"grant_type": "authorization_code",
		}).
		AcceptLang(acceptLang).
		EndBlob()

	if errs != nil {
		sugar.Error(errs)
		return wxlogin.AccessResponse{}, errs[0]
	}

	// {"access_token":"",
	// "expires_in":7200,
	// "refresh_token":"",
	// "openid":"ofP-k1LSVS-ObmrySM1aXKbv1Hjs",
	// "scope":"snsapi_login",
	// "unionid":"ogfvwjk6bFqv2yQpOrac0J3PqA0o"}
	sugar.Infof("GetAccessToken response: %s\n", resp.Body)

	var acc wxlogin.AccessResponse
	if err := json.Unmarshal(resp.Body, &acc); err != nil {
		sugar.Error(err)
		return acc, err
	}

	return acc, nil
}

// RefreshAccess refresh access token.
// See https://developers.weixin.qq.com/doc/oplatform/Mobile_App/WeChat_Login/Authorized_API_call_UnionID.html.
// access_token 是调用授权关系接口的调用凭证，由于 access_token 有效期（目前为 2 个小时）较短，当 access_token 超时后，可以使用 refresh_token 进行刷新，access_token 刷新结果有两种：
// 1.若 access_token 已超时，那么进行 refresh_token 会获取一个新的 access_token，新的超时时间；
// 2.若 access_token 未超时，那么进行 refresh_token 不会改变 access_token，但超时时间会刷新，相当于续期 access_token。
//
// refresh_token 拥有较长的有效期（30 天）且无法续期，当 refresh_token 失效的后，需要用户重新授权后才可以继续获取用户头像昵称。
func (c Client) RefreshAccess(refreshToken string) (wxlogin.AccessResponse, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	var acc wxlogin.AccessResponse
	resp, errs := fetch.New().
		Get(apiBaseURL + "/oauth2/refresh_token").
		SetQueryN(map[string]string{
			"appid":         c.app.AppID,
			"grant_type":    "refresh_token",
			"refresh_token": refreshToken,
		}).
		AcceptLang(acceptLang).
		EndBlob()

	sugar.Infof("RefreshAccess response: %s", resp.Body)

	if errs != nil {
		sugar.Error(errs)
		return acc, errs[0]
	}

	if err := json.Unmarshal(resp.Body, &acc); err != nil {
		sugar.Error(err)
		return acc, err
	}

	return acc, nil
}

// IsAccessTokenValid checks if an access token is valid.
func (c Client) IsAccessTokenValid(p wxlogin.UserInfoParams) bool {

	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	resp, errs := fetch.New().
		Get(apiBaseURL + "/auth").
		SetQueryN(map[string]string{
			"access_token": p.AccessToken,
			"openid":       p.OpenID,
		}).
		AcceptLang(acceptLang).
		EndBlob()

	if errs != nil {
		sugar.Error(errs)
		return false
	}

	sugar.Infof("IsAccessTokenValid response: %s", resp.Body)

	var rs wxlogin.RespStatus
	if err := json.Unmarshal(resp.Body, &rs); err != nil {
		sugar.Error(err)
		return false
	}

	if rs.Code == 0 && rs.Message == "ok" {
		return true
	}

	return false
}

// GetUserInfo from Wechat by open id.
// It seems wechat return empty fields as empty string.
func (c Client) GetUserInfo(p wxlogin.UserInfoParams) (wxlogin.UserInfoResponse, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	resp, errs := fetch.New().
		Get(apiBaseURL + "/userinfo").
		SetQueryN(map[string]string{
			"access_token": p.AccessToken,
			"openid":       p.OpenID,
		}).
		AcceptLang(acceptLang).
		EndBlob()

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

	if errs != nil {
		sugar.Error(errs)
		return wxlogin.UserInfoResponse{}, errs[0]
	}

	sugar.Infof("GetUserInfo response %s", resp.Body)

	var info wxlogin.UserInfoResponse
	if err := json.Unmarshal(resp.Body, &info); err != nil {
		sugar.Error(errs)
		return info, err
	}

	return info, nil
}
