package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/guregu/null"

	"gitlab.com/ftchinese/subscription-api/postoffice"
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
	apps    map[string]wxlogin.WxApp
	env     wxlogin.Env
	postman postoffice.PostMan
}

// NewWxAuth creates a new WxLoginRouter instance.
func NewWxAuth(db *sql.DB) WxAuthRouter {

	return WxAuthRouter{
		apps: wxlogin.Apps,
		env:  wxlogin.Env{DB: db},
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
		r := acc.BuildReason()
		view.Render(w, view.NewUnprocessable(r))
		return
	}

	client := util.NewClientApp(req)

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
		r := user.BuildReason()
		view.Render(w, view.NewUnprocessable(r))
		return
	}

	// Step 3:
	// Save access token
	go router.env.SaveAccess(app.AppID, acc, client)
	// Step 4:
	// Save userinfo
	err = router.env.SaveUserInfo(user)

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

	acc, err := router.env.LoadAccess(appID, sessionID)
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
			r := user.BuildReason()
			view.Render(w, view.NewUnprocessable(r))
			return
		}

		// Update wechat userinfo for this union id.
		err = router.env.UpdateUserInfo(user)

		if err != nil {
			view.Render(w, view.NewDBFailure(err))

			return
		}

		// 204 indicates user info is updated successfully.
		// Client can now request the updated account.
		view.Render(w, view.NewNoContent())
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
		r := user.BuildReason()
		view.Render(w, view.NewUnprocessable(r))
		return
	}

	// Save access token
	go router.env.UpdateAccess(sessionID, refreshedAcc.AccessToken)

	// Save userinfo
	err = router.env.UpdateUserInfo(user)

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

// LoadAccount gets a user's account data who logged in via wechat.
func (router WxAuthRouter) LoadAccount(w http.ResponseWriter, req *http.Request) {
	unionID := req.Header.Get(unionIDKey)

	account, err := router.env.FindAccountByWx(unionID)

	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	view.Render(w, view.NewResponse().NoCache().SetBody(account))
}

// BindAccount binds a FTC account to wechat.
// Binding accounts could be split into two step:
// 1. Add wechat union id to userinfo.wx_union_id.
// 2. Fill ftc_vip.vip_id and ftc_vip.vip_id_alias with ftc's account id and wechat's account id, if user purchased membership via either ftc account of wechat account.
func (router WxAuthRouter) BindAccount(w http.ResponseWriter, req *http.Request) {
	unionID := req.Header.Get(unionIDKey)

	userID, err := util.GetJSONString(req.Body, "userId")

	if err != nil {
		view.Render(w, view.NewBadRequest(""))
		return
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		reason := view.NewReason()
		reason.Field = "userId"
		reason.Field = view.CodeMissingField
		view.Render(w, view.NewUnprocessable(reason))
		return
	}

	// Find FTC account for this userID
	ftcAcnt, err := router.env.FindAccountByFTC(userID)
	// If the account if not found, deny the request -- you have nothing to bind.
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// Both ftcAcnt and wxAcnt should be found.
	// Otherwise how do you bind them?
	wxAcnt, err := router.env.FindAccountByWx(unionID)
	// If the wechat account if not found, deny the request -- you have nothing to bind to.
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// The two account already bound.
	if ftcAcnt.IsEqualTo(wxAcnt) {
		view.Render(w, view.NewNoContent())
		return
	}

	// If ftcAcnt is not equal to wxAcnt, the two accounts are separate accounts.
	// They might be bound to a 3rd account, or might not.
	// If any of them is bound to a 3rd account, deny the binding request.
	if ftcAcnt.IsCoupled() || wxAcnt.IsCoupled() {
		view.Render(w, view.NewForbidden("One of the requested accounts, or both, is/are bound to a 3rd account"))
		return
	}

	// Both accounts are not bound to any 3rd account. They are clear and binding is allowed to continue.
	// Next check memberships.

	// If both accounts have no memberships, simply set the userinfo.wx_union_id column to unionId.
	if !ftcAcnt.IsMember() && !wxAcnt.IsMember() {
		err := router.env.BindAccount(userID, unionID)

		if err != nil {
			view.Render(w, view.NewDBFailure(err))

			return
		}

		view.Render(w, view.NewNoContent())

		return
	}

	// If both accounts have memberships.
	if ftcAcnt.IsMember() && wxAcnt.IsMember() {
		// If the two accounts' memberships point to the same one, just bind account and ignore membership binding.
		if ftcAcnt.Membership.IsEqualTo(wxAcnt.Membership) {
			err := router.env.BindAccount(userID, unionID)

			if err != nil {
				view.Render(w, view.NewDBFailure(err))

				return
			}

			view.Render(w, view.NewNoContent())

			return
		}

		// The memberships of both accounts are separate ones.
		// If any of two accounts is bound to a 3rd account, this 3rd account must be different from the two.
		// You should not merge the membership.
		if ftcAcnt.Membership.IsCoupled() || wxAcnt.Membership.IsCoupled() {
			view.Render(w, view.NewForbidden("The membership of one of the requested accounts, or both, is/are bound to a 3rd account"))
			return
		}

		// None of the two accounts' membership are bound to a 3rd accounts.
		// You might be able to merge the memberships.
		// If both membership are valid, deny request.
		if !ftcAcnt.Membership.IsExpired() && !wxAcnt.Membership.IsExpired() {
			view.Render(w, view.NewForbidden("The two accounts have different valid memberships!"))
		}

		// If both membership are invalid, or one of the memberships are invalid, you can merge them.
		// First merge the two memberships.
		// Then perform db operation:
		// save the membership to be deleted in another table;
		// bind account;
		// delete wechat membership;
		// insert or update ftc membership
		merged := ftcAcnt.Membership.Merge(wxAcnt.Membership)

		// Save the wechat membership to another table.
		go router.env.SaveMergedMember(userID, wxAcnt.Membership)

		err := router.env.BindAccountAndMember(merged)

		if err != nil {
			view.Render(w, view.NewDBFailure(err))
			return
		}

		emailBody := wxlogin.EmailBody{
			UserName:   ftcAcnt.UserName.String,
			Email:      ftcAcnt.Email,
			NickName:   wxAcnt.Wechat.NickName,
			MutexMerge: !ftcAcnt.Membership.IsExpired() || !wxAcnt.Membership.IsExpired(),
			FTCMember:  ftcAcnt.Membership,
			WxMember:   wxAcnt.Membership,
		}

		go router.sendBoundLetter(emailBody)

		view.Render(w, view.NewNoContent())
		return
	}

	// Only one of the accounts has membership.
	// Bind accounts.
	// Create or update entry in ftc_vip
	// Update the valid membership
	picked := ftcAcnt.Membership.Pick(wxAcnt.Membership)
	picked.UserID = userID
	picked.UnionID = null.StringFrom(unionID)

	err = router.env.BindAccountAndMember(picked)

	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	emailBody := wxlogin.EmailBody{
		UserName:   ftcAcnt.UserName.String,
		Email:      ftcAcnt.Email,
		NickName:   wxAcnt.Wechat.NickName,
		MutexMerge: false,
		FTCMember:  ftcAcnt.Membership,
		WxMember:   wxAcnt.Membership,
	}

	go router.sendBoundLetter(emailBody)

	view.Render(w, view.NewNoContent())
}

func (router WxAuthRouter) sendBoundLetter(e wxlogin.EmailBody) error {
	p, err := e.ComposeParcel()
	if err != nil {
		return err
	}
	return router.postman.Deliver(p)
}
