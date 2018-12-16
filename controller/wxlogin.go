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

// BindFTC binds a FTC account to wechat.
// Binding accounts could be split into two step:
// 1. Add wechat union id to userinfo.wx_union_id.
// 2. Fill ftc_vip.vip_id and ftc_vip.vip_id_alias with ftc's account id and wechat's account id, if user purchased membership via either ftc account of wechat account.
func (lr WxAuthRouter) BindFTC(w http.ResponseWriter, req *http.Request) {
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
	ftcAcnt, err := lr.env.CheckFTCAccount(userID)
	// If the account if not found, deny the request -- you have nothing to bind.
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// Both ftcAcnt and wxAcnt should be found.
	// Otherwise how do you bind them?
	wxAcnt, err := lr.env.CheckWxAccount(unionID)
	// If the wechat account if not found, deny the request -- you have nothing to bind to.
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// If this ftc account is already bound to a wechat account
	if ftcAcnt.UnionID.Valid {
		// If the ftc account is already bound to another wechat account.
		if ftcAcnt.UnionID.String != unionID {
			view.Render(w, view.NewForbidden("The ftc acount is bound to another wechat account"))
			return
		}

		// Check if membership is correctly bound.
		// If none of the accounts have memberships, then we do not need to be borthered with ftc_vip table.
		if !ftcAcnt.IsMember && !wxAcnt.IsMember {
			view.Render(w, view.NewNoContent())
			return
		}

		// Here it means, both accounts have membership, or one of them does.
		// If both accounts have membership, their vip_id_alias should be the same.
		if ftcAcnt.IsMember && wxAcnt.IsMember {
			// If both ftc account and wx account have memberships,
			// their MemberUnionID should be the same,
			// otherwise there are problems.
			if ftcAcnt.MemberUnionID.String != unionID {
				view.Render(w, view.NewForbidden("The ftc acount membership is bound to another wechat account"))
			}

			view.Render(w, view.NewNoContent())
			return
		}

		// If either ftc account or wechat account has membership.
		bound := wxlogin.BoundAccount{
			UserID:  userID,
			UnionID: unionID,
		}
		if ftcAcnt.IsMember {
			bound.Method = wxlogin.MethodEmail
		} else if wxAcnt.IsMember {
			bound.Method = wxlogin.MethodWx
		} else {
			bound.Method = wxlogin.MethodNone
		}

		err := lr.env.MergeMembership(bound)

		if err != nil {
			view.Render(w, view.NewDBFailure(err))
			return
		}

		view.Render(w, view.NewNoContent())
		return
	}

	// If both the ftc account and wechat account has their membership, do nothing.
	if ftcAcnt.IsMember && wxAcnt.IsMember {
		// ftcAcnt and wxAcnt might retrieve the same membership due to legacy issues.
		if ftcAcnt.MemberUnionID.Valid && wxAcnt.MemberUnionID.Valid {
			// This means membership is bound but accounts are not.
			if ftcAcnt.MemberUnionID.String == wxAcnt.MemberUnionID.String {
				err := lr.env.BindAccount(userID, unionID)
				if err != nil {
					view.Render(w, view.NewDBFailure(err))
					return
				}
				view.Render(w, view.NewNoContent())
				return
			}
		}
		view.Render(w, view.NewForbidden("Refuse to merge two accounts with subscribed memberships"))
		return
	}

	// If none of ftc account nor wechat account has membership.
	if !ftcAcnt.IsMember && !wxAcnt.IsMember {
		err := lr.env.BindAccount(userID, unionID)

		if err != nil {
			view.Render(w, view.NewDBFailure(err))
			return
		}

		view.Render(w, view.NewNoContent())
		return
	}

	// If only one of wechat or ftc account has membership attached,
	// bind two accounts and merge membership.
	bound := wxlogin.BoundAccount{
		UserID:  userID,
		UnionID: unionID,
	}
	if ftcAcnt.IsMember {
		bound.Method = wxlogin.MethodEmail
	} else if wxAcnt.IsMember {
		bound.Method = wxlogin.MethodWx
	} else {
		bound.Method = wxlogin.MethodNone
	}

	err = lr.env.MergeAccount(bound)

	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	view.Render(w, view.NewNoContent())
}
