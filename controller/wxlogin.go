package controller

import (
	"net/http"
	"strings"

	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/view"
	"gitlab.com/ftchinese/subscription-api/wxlogin"
)

// WxLoginRouter handles wechat login.
type WxLoginRouter struct {
	client wxlogin.Client
	env    wxlogin.Env
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
func (lr WxLoginRouter) Login(w http.ResponseWriter, req *http.Request) {
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

	// Request access token from wechat
	acc, err := lr.client.GetAccessToken(code)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))

		return
	}

	reqClient := util.NewRequestClient(req)

	// Save access token
	go lr.env.SaveAccess(acc, reqClient)

	// Get userinfo from wechat
	user, err := lr.client.GetUserInfo(acc)

	// Save userinfo
	go lr.env.SaveUserInfo(user, reqClient)

	// TODO: change to cucurrent retrieval later.
	var acnt wxlogin.Account
	acnt, _ = lr.env.FindAccount(acc.UnionID)

	member, err := lr.env.FindMembership(acc.UnionID)

	acnt.Membership = member

	acnt.Wechat = &wxlogin.WxAccount{
		OpenID:    user.OpenID,
		NickName:  user.NickName,
		AvatarURL: user.HeadImgURL,
		UnionID:   user.UnionID,
	}

	view.Render(w, view.NewResponse().NoCache().SetBody(acnt))
}
