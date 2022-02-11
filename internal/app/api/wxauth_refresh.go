package api

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"net/http"
)

// Refresh allows user to refresh userinfo.
// Request header must contain `X-App-Id`.
// Request body
// * sessionId: string
//
// Error
// 422: refresh_token_invalid
func (router WxAuthRouter) Refresh(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	wxClient, err := router.getClient(req)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	var param input.WechatRefreshParams
	if err := gorest.ParseJSON(req.Body, &param); err != nil {
		sugar.Error(err)
		_ = render.NewBadRequest(err.Error())
		return
	}

	if ve := param.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.NewUnprocessable(ve)
		return
	}

	// Load previously saved access token.
	currentAccessSchema, err := router.wxRepo.LoadWxAccess(
		wxClient.GetApp().AppID,
		param.SessionID)
	// Access token for this openID + appID + clientType is not found
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// 检验授权凭证（access_token）是否有效
	isValid := wxClient.IsAccessTokenValid(
		currentAccessSchema.UserInfoParams())

	// 刷新或续期 access_token 使用
	if !isValid {
		sugar.Infof("Access token %s is no longer valid", currentAccessSchema.AccessToken)

		// Access token is no longer valid. Refresh access token.
		// Get the new access token.
		newAccessResp, err := wxClient.RefreshAccess(
			currentAccessSchema.RefreshToken)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).BadRequest(err.Error())
			return
		}

		// Handle response error.
		// Caused by: invalid refresh token.
		// Client should ask infoResp to re-authorize
		if newAccessResp.HasError() {
			router.handleWxApiErr(w, newAccessResp.RespStatus)
			return
		}

		// Refresh current access token schema.
		currentAccessSchema = currentAccessSchema.WithAccessToken(newAccessResp.AccessToken)
		// Update access token.
		go func() {
			_ = router.wxRepo.UpdateWxAccess(currentAccessSchema)
		}()
	}

	_, err = router.getUserInfo(
		w,
		wxClient,
		currentAccessSchema.UserInfoParams(),
	)

	if err != nil {
		return
	}

	_ = render.New(w).NoContent()
}
