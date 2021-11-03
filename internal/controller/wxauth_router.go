package controller

import (
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/wxoauth"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"go.uber.org/zap"
	"net/http"
)

// WxAuthRouter handles wechat login.
// Web apps and mobile apps should use their
// respective app id + app secret combination.
// Wechat never said you should do this.
// But when combining their messy documentation, you must do it this way.
type WxAuthRouter struct {
	apps   map[string]config.WechatApp
	wxRepo wxoauth.Env
	logger *zap.Logger
}

// NewWxAuth creates a new WxLoginRouter instance.
func NewWxAuth(dbs db.ReadWriteMyDBs, logger *zap.Logger) WxAuthRouter {
	return WxAuthRouter{
		apps:   config.MustGetWechatApps(),
		wxRepo: wxoauth.NewEnv(dbs),
		logger: logger,
	}
}

func (router WxAuthRouter) getClient(req *http.Request) (wxoauth.Client, error) {
	// Find this wxClient.
	appID := req.Header.Get("X-App-Id")
	app, ok := router.apps[appID]
	if !ok {
		return wxoauth.Client{}, errors.New("unknown app")
	}

	return wxoauth.NewClient(app), nil
}

func (router WxAuthRouter) handleWxApiErr(w http.ResponseWriter, rs wxlogin.RespStatus) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	sugar.Error(rs.Message)

	go func() {
		err := router.wxRepo.SaveWxStatus(rs)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).Unprocessable(rs.GetInvalidity())
}

func (router WxAuthRouter) getUserInfo(w http.ResponseWriter, client wxoauth.Client, param wxlogin.UserInfoParams) (wxlogin.UserInfoSchema, error) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Use access token to get userinfo from wechat.
	infoResp, err := client.GetUserInfo(
		param)

	// Request has error.
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return wxlogin.UserInfoSchema{}, err
	}

	// Handle Wechat response error.
	// Cause by: invalid access token, invalid open id.
	// Just ask infoResp to retry.
	if infoResp.HasError() {
		router.handleWxApiErr(w, infoResp.RespStatus)
		return wxlogin.UserInfoSchema{}, err
	}

	infoSchema := wxlogin.NewUserInfo(infoResp)

	// Save userinfo
	err = router.wxRepo.UpsertUserInfo(
		wxlogin.NewUserInfo(infoResp))

	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return wxlogin.UserInfoSchema{}, err
	}

	return infoSchema, nil
}
