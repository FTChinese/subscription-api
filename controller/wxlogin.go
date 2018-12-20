package controller

import (
	"database/sql"
	"net/http"
	"os"
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
	// wClient is used to handle web app login request
	wClient wxlogin.Client
	// mClient is used to handle mobile app login request.
	mClient wxlogin.Client
	env     wxlogin.Env
	postman postoffice.PostMan
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
		reason.Code = view.CodeMissingField
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

	view.Render(w, view.NewResponse().NoCache().SetBody(user.ToWechat()))
}

// LoadAccount gets a user's account data who logged in via wechat.
func (lr WxAuthRouter) LoadAccount(w http.ResponseWriter, req *http.Request) {
	unionID := req.Header.Get(unionIDKey)

	account, err := lr.env.FindAccountByWx(unionID)

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
func (lr WxAuthRouter) BindAccount(w http.ResponseWriter, req *http.Request) {
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
	ftcAcnt, err := lr.env.FindAccountByFTC(userID)
	// If the account if not found, deny the request -- you have nothing to bind.
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// Both ftcAcnt and wxAcnt should be found.
	// Otherwise how do you bind them?
	wxAcnt, err := lr.env.FindAccountByWx(unionID)
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
		err := lr.env.BindAccount(userID, unionID)

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
			err := lr.env.BindAccount(userID, unionID)

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
		go lr.env.SaveMergedMember(userID, wxAcnt.Membership)

		err := lr.env.BindAccountAndMember(merged)

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

		go lr.sendBoundLetter(emailBody)

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

	err = lr.env.BindAccountAndMember(picked)

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

	go lr.sendBoundLetter(emailBody)

	view.Render(w, view.NewNoContent())
}

func (lr WxAuthRouter) sendBoundLetter(e wxlogin.EmailBody) error {
	p, err := e.ComposeParcel()
	if err != nil {
		return err
	}
	return lr.postman.Deliver(p)
}
