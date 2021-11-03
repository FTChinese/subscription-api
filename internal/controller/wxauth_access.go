package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"net/http"
)

// Login performs the Step 2 of OAuth as
// described by https://open.weixin.qq.com/cgi-bin/showdocument?action=dir_list&t=resource/res_list&verify=1&id=open1419317851&token=&lang=zh_CN.
//
// It uses Wechat's OAuth code to exchange for access token, and then use access token to get user info.
//
// The code is acquired by different approach depending on the platform:
// * For native app, it gets the code by calling Wechat SDK;
// * For web app, it sends a GET request to Wechat API,
// wechat redirect to this API's callback endpoint,
// and this api redirect back to the web app's callback url.
//
// After getting the code, client app send the code here.
// Client should also include the app id issued by Wechat which it used to apply for the code.
// Since the code is bound to the app id, this API must know which app id to use to perform the following steps.
// Use the `X-App-Id` key in request header.
// Returns a Session wxlogin.Session instance.
// Client could then fetch user account from /account/wx endpoint by setting request header X-Union-Id.
//
// Request body:
// * code: xxxxx
//
// Request header:
// * X-App-Id: xxxx
//
// Error:
// 422:
// * code_missing_field;
// * code_invalid;
// * openId_invalid
func (router WxAuthRouter) Login(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	wxClient, err := router.getClient(req)
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	sugar.Infof("Using wechat app: %v", wxClient.GetApp())

	var param input.WechatAccessParams
	if err := gorest.ParseJSON(req.Body, &param); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := param.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// 通过code参数加上AppID和AppSecret等，通过API换取access_token
	// Exchange access token with code.
	// Error only indicates network failure.
	// Wechat error is still a 200 OK response.
	tokenResp, err := wxClient.GetAccessToken(param.Code)
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Handle wechat response error.
	if tokenResp.HasError() {
		router.handleWxApiErr(w, tokenResp.RespStatus)
		return
	}

	tokenSchema := wxlogin.NewAccessSchema(
		tokenResp,
		wxClient.GetApp().AppID,
		footprint.NewClient(req))

	// Save access token
	go func() {
		if err := router.wxRepo.SaveWxAccess(tokenSchema); err != nil {
			sugar.Error(err)
		}
	}()

	// 3. 通过access_token进行接口调用，获取用户基本数据资源或帮助用户实现基本操作。
	infoSchema, err := router.getUserInfo(
		w,
		wxClient,
		tokenResp.UserInfoParams())

	if err != nil {
		return
	}

	// Send session data to client.
	_ = render.New(w).OK(
		wxlogin.NewSession(tokenSchema, infoSchema.UnionID))
}
