package api

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

type AccountRouter struct {
	UserShared
}

func NewAccountRouter(shared UserShared) AccountRouter {
	return AccountRouter{
		UserShared: shared,
	}
}

// LoadAccountByFtcID loads a user's full account data
// by ftc id provided in request header.
func (router AccountRouter) LoadAccountByFtcID(w http.ResponseWriter, req *http.Request) {
	userID := xhttp.GetFtcID(req.Header)

	acnt, err := router.ReaderRepo.AccountByFtcID(userID)

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
	unionID := xhttp.GetUnionID(req.Header)

	acnt, err := router.ReaderRepo.AccountByWxID(unionID)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(acnt)
}

// DeleteFtcAccount verifies user credentials and delete its account.
// Input
// * email: string;
// * password: string.
//
// Deletion is not permitted if any of the following conditions is not met:
// * Password must be correct for this id - Status Forbidden
// * Email must match the email under this id - Unprocessable email_missing.
// * This id should not have a valid subscription - Unprocessable subscription_already_exists.
func (router AccountRouter) DeleteFtcAccount(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	userID := xhttp.GetFtcID(req.Header)

	var params input.EmailCredentials
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
	}

	if ve := params.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Verify current password.
	authResult, err := router.Repo.VerifyIDPassword(account.IDCredentials{
		FtcID:    userID,
		Password: params.Password,
	})
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Password incorrect.
	if !authResult.PasswordMatched {
		_ = render.New(w).Forbidden("Current password incorrect")
		return
	}

	// Password verified.
	acnt, err := router.ReaderRepo.AccountByFtcID(userID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// email_missing: the requested email does not match this account's email, thus resource missing.
	// subscription_already_exists: a valid membership exists, thus deletion not allowed.
	if ve := acnt.VerifyDelete(params.Email); ve != nil {
		sugar.Error(err)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	err = router.Repo.DeleteAccount(acnt.Deleted())
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).NoContent()
}
