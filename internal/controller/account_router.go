package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
	"net/http"
)

type AccountRouter struct {
	UserShared
}

func NewAccountRouter(myDBs db.ReadWriteMyDBs, postman postoffice.PostOffice, l *zap.Logger) AccountRouter {

	return AccountRouter{
		UserShared: NewUserShared(myDBs, postman, l),
	}
}

// LoadAccountByFtcID loads a user's full account data
// by ftc id provided in request header.
func (router AccountRouter) LoadAccountByFtcID(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get(ftcIDKey)

	acnt, err := router.userRepo.AccountByFtcID(userID)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if acnt.IsMobileEmail() {
		acnt.BaseAccount = acnt.SyncMobile()
		acnt.LoginMethod = enum.LoginMethodMobile
		router.SyncMobile(acnt.BaseAccount)
	}

	_ = render.New(w).OK(acnt)
}

// LoadAccountByWx respond to request for user account by X-Union-Id.
//
//	GET /wx/account
// Header `X-Union-Id: <wechat union id>`
func (router AccountRouter) LoadAccountByWx(w http.ResponseWriter, req *http.Request) {
	unionID := req.Header.Get(unionIDKey)

	acnt, err := router.userRepo.AccountByWxID(unionID)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(acnt)
}
