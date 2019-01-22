package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/view"
	cache "github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/wxlogin"
)

// WxAuthRouter handles wechat login.
// Web apps and mobile apps should use their
// respective app id + app secret combination.
// Wechat never said you should do this.
// But when combining their messy documentation, you must do it this way.
type WxAuthRouter struct {
	apps  map[string]wxlogin.WxApp
	model model.Env
}

// NewWxAuth creates a new WxLoginRouter instance.
func NewWxAuth(db *sql.DB, c *cache.Cache) WxAuthRouter {

	// 移动应用 -> FT中文网会员订阅. This is used for Android subscription
	mSubs, err := wxlogin.NewWxApp(
		os.Getenv("WXPAY_APPID"),
		os.Getenv("WXPAY_APPSECRET"),
	)
	if err != nil {
		os.Exit(1)
	}
	// 移动应用 -> FT中文网. This is for iOS subscription and legacy Android subscription.
	mFTC, err := wxlogin.NewWxApp(
		os.Getenv("WX_MOBILE_APPID"),
		os.Getenv("WX_MOBILE_APPSECRET"),
	)
	if err != nil {
		os.Exit(1)
	}
	// 网站应用 -> FT中文网. This is used for web login
	wFTC, err := wxlogin.NewWxApp(
		os.Getenv("wxc7233549ca6bc86a"),
		os.Getenv("098330adf494c46d368868a799320a4e"),
	)
	if err != nil {
		os.Exit(1)
	}

	return WxAuthRouter{
		apps: map[string]wxlogin.WxApp{
			// 移动应用 -> FT中文网会员订阅. This is used for Android subscription
			"wxacddf1c20516eb69": mSubs,
			// 移动应用 -> FT中文网. This is for iOS subscription and legacy Android subscription.
			"wxc1bc20ee7478536a": mFTC,
			// 网站应用 -> FT中文网. This is used for web login
			"wxc7233549ca6bc86a": wFTC,
		},
		model: model.New(db, c, false),
	}
}

// Login performs the Step 2 of OAuth as
// described by https://open.weixin.qq.com/cgi-bin/showdocument?action=dir_list&t=resource/res_list&verify=1&id=open1419317851&token=&lang=zh_CN.
//
// It uses Wechat's OAuth code to exchange for access token, and then use access token to get user info.
//
// Input {code: "oauth code"}.
//
// For native app, it gets the code by calling Wechat SDK;
// For web app, it sends a GET request to Wechat API,
// wechat redirect this this API's callback endpoint,
// and this api redirect back to the web app's callback url.
//
// After getting the code, client app send the code here.
// Client should also include the app id issued by Wechat which it used to apply for the code.
// Since the code is bound to the app id, this API must know which which app id to use to perform the folowing steps.
// Use the `X-App-Id` key in request header.
func (router WxAuthRouter) Login(w http.ResponseWriter, req *http.Request) {
	appID := req.Header.Get("X-App-Id")
	app, ok := router.apps[appID]
	if !ok {
		view.Render(w, view.NewBadRequest("Unknown app"))
		return
	}

	logger.WithField("trace", "Login").Infof("Wechat app: %+v", app)

	// Get `code` from request body
	code, err := util.GetJSONString(req.Body, "code")

	if err != nil {
		view.Render(w, view.NewBadRequest(""))
		return
	}
	// Make sure `code` exists.
	code = strings.TrimSpace(code)
	if code == "" {
		reason := view.NewReason()
		reason.Field = "code"
		reason.Code = view.CodeMissingField
		view.Render(w, view.NewUnprocessable(reason))
		return
	}

	// Step 1:
	// Exchange access token with code.
	// Error only indicates network failure.
	// Wechat error is still a 200 OK response.
	acc, err := app.GetAccessToken(code)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))

		return
	}

	// Handle wechat response error.
	if acc.HasError() {
		logger.WithField("trace", "Login GetAccessToken").Error(acc.Message)

		// Log Wechat error response
		go router.model.SaveWxStatus(acc.Code, acc.Message)

		r := acc.BuildReason()
		view.Render(w, view.NewUnprocessable(r))
		return
	}

	client := gorest.NewClientApp(req)

	// Step 2:
	// Use access token to get userinfo from wechat
	user, err := app.GetUserInfo(acc.AccessToken, acc.OpenID)
	// Request has error.
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))

		return
	}

	// Handle Wechat response error.
	// Cause by: invalid access token, invalid open id.
	// Just ask user to retry.
	if user.HasError() {
		// Log error response.
		go router.model.SaveWxStatus(user.Code, user.Message)

		r := user.BuildReason()
		view.Render(w, view.NewUnprocessable(r))
		return
	}

	// Step 3:
	// Save access token
	go router.model.SaveWxAccess(appID, acc, client)

	// Step 4:
	// Save userinfo
	err = router.model.SaveWxUser(user)

	if err != nil {
		view.Render(w, view.NewDBFailure(err))

		return
	}

	// Send session data to client.
	view.Render(w, view.NewResponse().NoCache().SetBody(acc.ToSession(user.UnionID)))
}

// Refresh allows user to refresh userinfo.
// Request header must contain `X-App-Id`.
// Input {openId: string}
func (router WxAuthRouter) Refresh(w http.ResponseWriter, req *http.Request) {
	appID := req.Header.Get("X-App-Id")
	app, ok := router.apps[appID]
	if !ok {
		view.Render(w, view.NewBadRequest("Unknown app"))
		return
	}

	// Parse request body
	sessionID, err := util.GetJSONString(req.Body, "sessionId")

	acc, err := router.model.LoadWxAccess(appID, sessionID)
	// Access token for this openID + appID + clientType is not found
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	isValid := app.IsValidAccess(acc.AccessToken, acc.OpenID)

	// If access token is still valid.
	if isValid {
		// Use access token to get userinfo from wechat
		user, err := app.GetUserInfo(acc.AccessToken, acc.OpenID)
		// Request has error.
		if err != nil {
			view.Render(w, view.NewBadRequest(err.Error()))

			return
		}

		// Handle Wechat response error.
		// Cause by: invalid access token, invalid open id.
		// Just ask user to retry.
		if user.HasError() {
			go router.model.SaveWxStatus(user.Code, user.Message)

			r := user.BuildReason()
			view.Render(w, view.NewUnprocessable(r))
			return
		}

		// Update wechat userinfo for this union id.
		err = router.model.UpdateWxUser(user)

		if err != nil {
			view.Render(w, view.NewDBFailure(err))

			return
		}

		// 204 indicates user info is updated successfully.
		// Client can now request the updated account.
		view.Render(w, view.NewNoContent())
		return
	}

	// Access token is no longer valid. Refresh access token
	refreshedAcc, err := app.RefreshAccess(acc.RefreshToken)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// Handle wechat response error.
	// Caused by: invalid refresh token.
	if acc.HasError() {
		go router.model.SaveWxStatus(acc.Code, acc.Message)

		r := acc.BuildReason()
		view.Render(w, view.NewUnprocessable(r))
		return
	}

	// Use access token to get userinfo from wechat
	user, err := app.GetUserInfo(acc.AccessToken, acc.OpenID)
	// Request has error.
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))

		return
	}

	// Handle Wechat response error.
	// Cause by: invalid access token, invalid open id.
	// Just ask user to retry.
	if user.HasError() {
		go router.model.SaveWxStatus(user.Code, user.Message)
		r := user.BuildReason()
		view.Render(w, view.NewUnprocessable(r))
		return
	}

	// Save access token
	go router.model.UpdateWxAccess(sessionID, refreshedAcc.AccessToken)

	// Save userinfo
	err = router.model.UpdateWxUser(user)

	if err != nil {
		view.Render(w, view.NewDBFailure(err))

		return
	}

	view.Render(w, view.NewNoContent())
}

// WebCallback is used to help web app to get OAuth 2.0 code.
// The code and state is transferred back to next-user web app since Wechat only recognize the ftacacemy.cn URL.
func (router WxAuthRouter) WebCallback(w http.ResponseWriter, req *http.Request) {
	// The code returned by wechat
	code := req.FormValue("code")
	// The nonce code we send to wechat.
	state := req.FormValue("state")

	code = strings.TrimSpace(code)

	if code == "" {
		view.Render(w, view.NewForbidden("Authorization denied"))
		return
	}

	http.Redirect(w, req, fmt.Sprintf("http://localhost:4100/callback?code=%s&state=%s", code, state), http.StatusSeeOther)
}
