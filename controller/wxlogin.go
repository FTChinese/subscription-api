package controller

import (
	"database/sql"
	"net/http"
	"os"
	"strings"

	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/view"
	"gitlab.com/ftchinese/subscription-api/wxlogin"
)

// WxAuthRouter handles wechat login.
// Web apps and mobile apps should use their
// respective app id + app secret combination.
// Wechat never said you should do this.
// But when combining their messy documentation, you must do it this way.
type WxAuthRouter struct {
	// wClient is used to handle web app login request
	wClient wxlogin.Client
	// mClient is used to handle mobile app login request.
	mClient wxlogin.Client
	env     wxlogin.Env
}

// NewWxAuth creates a new WxLoginRouter instance.
func NewWxAuth(db *sql.DB) WxAuthRouter {
	mAppID := os.Getenv("WX_MOBILE_APPID")
	mAppScrt := os.Getenv("WX_MOBILE_APPSECRET")

	wAppID := os.Getenv("WX_WEB_APPID")
	wAppScrt := os.Getenv("WX_WEB_APPSECRET")

	return WxAuthRouter{
		wClient: wxlogin.NewClient(wAppID, wAppScrt),
		mClient: wxlogin.NewClient(mAppID, mAppScrt),
		env:     wxlogin.Env{DB: db},
	}
}

// Login handles login via wechat.
// Input {code: "oauth code"}.
// Client send the oauth code it requested from
// Wechat API.
// Login performs the Step 2 of OAuth as
// described by https://open.weixin.qq.com/cgi-bin/showdocument?action=dir_list&t=resource/res_list&verify=1&id=open1419317851&token=&lang=zh_CN.
//
// It uses the code for exchange of access
// token, then save the access token.
// After that, it uses the access token and open id to get user info from wechat,
// send it back to client.
// After getting a user's wechat data,
// client should then retrieve the complete user data: FTC account + wechat userinfo + membership
func (lr WxAuthRouter) Login(w http.ResponseWriter, req *http.Request) {
	// Parse request body
	code, err := util.GetJSONString(req.Body, "code")

	if err != nil {
		view.Render(w, view.NewBadRequest(""))
		return
	}

	code = strings.TrimSpace(code)
	if code == "" {
		reason := view.NewReason()
		reason.Field = "code"
		reason.Field = view.CodeMissingField
		view.Render(w, view.NewUnprocessable(reason))
		return
	}

	// TODO: Use client type to determine which wechat app id will be used.
	reqClient := util.GetClient(req)

	// Request access token from wechat
	acc, err := lr.mClient.GetAccessToken(code)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))

		return
	}

	// Save access token
	go lr.env.SaveAccess(acc, reqClient)

	// Get userinfo from wechat
	user, err := lr.mClient.GetUserInfo(acc)

	// Save userinfo
	err = lr.env.SaveUserInfo(user, reqClient)

	if err != nil {
		view.Render(w, view.NewDBFailure(err))

		return
	}

	view.Render(w, view.NewResponse().NoCache().SetBody(user.WxAccount()))
}

// LoadAccount gets a user's account data who logged in via wechat.
func (lr WxAuthRouter) LoadAccount(w http.ResponseWriter, req *http.Request) {
	unionID := req.Header.Get(unionIDKey)

	account, err := lr.env.LoadAccountByWx(unionID)

	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	view.Render(w, view.NewResponse().NoCache().SetBody(account))
}
