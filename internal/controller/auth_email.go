package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"net/http"
)

// EmailExists tests if a email exists.
//
//	GET /auth/email/exists?v={email}
func (router AuthRouter) EmailExists(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	email := req.FormValue("v")

	if ve := validator.EnsureEmail(email); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	ok, err := router.repo.EmailExists(email)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if !ok {
		_ = render.New(w).NotFound("")
		return
	}

	_ = render.New(w).NoContent()
}

// EmailLogin handles user login with email + password.
//
// 	POST /users/login
//
// email: string,
// password: string
// deviceToken?: string. Required only for Android app.
func (router AuthRouter) EmailLogin(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var params pkg.EmailLoginParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	authResult, err := router.repo.Authenticate(params)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if !authResult.PasswordMatched {
		_ = render.New(w).Forbidden("Incorrect credentials")
		return
	}

	acnt, err := router.repo.AccountByFtcID(authResult.UserID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
	}

	fp := footprint.New(acnt.FtcID, footprint.NewClient(req)).
		FromLogin().
		WithAuth(enum.LoginMethodEmail, params.DeviceToken)

	go func() {
		err := router.repo.SaveFootprint(fp)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(acnt)
}

// EmailSignUp create a new account for a user.
//
// 	POST /users/signup
//
// Input:
// email: string
// password: string
// sourceUrl?: string From which site the request is sent.
//
// `sourceUrl` is used to build verification URL. Only required for desktop browsers.
// Mobile apps does not need to provide this field.
func (router AuthRouter) EmailSignUp(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	clientApp := footprint.NewClient(req)

	if !clientApp.UserIP.IsZero() {
		// SingupCount error should be omitted.
		// This should not be a barrier for signup process.
		limit, _ := router.repo.SignUpCount(account.NewSignUpRateParams(clientApp.UserIP.String, 1))

		if limit.Exceeds() {
			_ = render.New(w).TooManyRequests("Too many accounts were created at the same IP within the past 1 hour.")
			return
		}
	}

	var params pkg.EmailSignUpParams

	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	baseAccount := account.NewEmailBaseAccount(params)
	err := router.repo.CreateAccount(baseAccount)
	if err != nil {
		sugar.Error(err)
		if db.IsAlreadyExists(err) {
			_ = render.New(w).Unprocessable(render.NewVEAlreadyExists("email"))
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	fp := footprint.New(baseAccount.FtcID, clientApp).
		FromSignUp().
		WithAuth(enum.LoginMethodEmail, params.DeviceToken)

	go func() {
		err := router.repo.SaveFootprint(fp)
		if err != nil {
			sugar.Error(err)
		}
	}()

	go func() {
		_ = router.SendEmailVerification(
			baseAccount,
			params.SourceURL,
			true)
	}()

	_ = render.New(w).OK(reader.Account{
		BaseAccount: baseAccount,
		LoginMethod: enum.LoginMethodEmail,
		Wechat:      account.Wechat{},
		Membership:  reader.Membership{},
	})
}

// VerifyEmail verifies the authenticity of user's email.
//
//	POST /auth/email/verification/{token}
func (router AuthRouter) VerifyEmail(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	token, err := getURLParam(req, "token").ToString()
	// `400 Bad Request`
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	vrf, err := router.repo.RetrieveEmailVerifier(token)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	baseAccount, err := router.repo.BaseAccountByEmail(vrf.Email)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if baseAccount.IsVerified {
		_ = render.New(w).NoContent()
		return
	}

	if err := router.repo.EmailVerified(baseAccount.FtcID); err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	fp := footprint.
		New(baseAccount.FtcID, footprint.NewClient(req)).
		FromVerification()

	go func() {
		_ = router.repo.SaveFootprint(fp)
	}()

	// Send a greeting letter.
	go func() {
		parcel, err := letter.GreetingParcel(baseAccount)
		if err != nil {
			return
		}

		err = router.postman.Deliver(parcel)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).NoContent()
}