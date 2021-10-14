package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/lib/validator"
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

	ok, err := router.userRepo.EmailExists(email)
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
// Input: pkg.EmailLoginParams.
// * email: string,
// * password: string
// * deviceToken?: string. Required only for Android app.
//
// The footprint.Client headers are required.
func (router AuthRouter) EmailLogin(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var params input.EmailLoginParams
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

	// Not found if email does not exist
	authResult, err := router.userRepo.Authenticate(params.EmailCredentials)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Forbidden if password incorrect.
	if !authResult.PasswordMatched {
		_ = render.New(w).Forbidden("Incorrect credentials")
		return
	}

	// There shouldn't be any not found error.
	acnt, err := router.userRepo.AccountByFtcID(authResult.UserID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
	}

	fp := footprint.New(acnt.FtcID, footprint.NewClient(req)).
		FromLogin().
		WithAuth(enum.LoginMethodEmail, params.DeviceToken)

	go func() {
		err := router.userRepo.SaveFootprint(fp)
		if err != nil {
			sugar.Error(err)
		}
	}()

	if acnt.IsMobileEmail() {
		acnt.BaseAccount = acnt.SyncMobile()
		acnt.LoginMethod = enum.LoginMethodMobile
		router.SyncMobile(acnt.BaseAccount)
	}

	_ = render.New(w).OK(acnt)
}

// EmailSignUp create a new account for a user.
//
// 	POST /users/signup
//
// Input:
// * email: string;
// * password: string;
// * mobile?: string;
// * deviceToken?: string;
// * sourceUrl?: string From which site the request is sent. Not required for mobile apps.
//
// The footprint.Client headers are required.
func (router AuthRouter) EmailSignUp(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	clientApp := footprint.NewClient(req)

	if !clientApp.UserIP.IsZero() {
		// SingupCount error should be omitted.
		// This should not be a barrier for signup process.
		limit, _ := router.userRepo.SignUpCount(account.NewSignUpRateParams(clientApp.UserIP.String, 1))

		if limit.Exceeds() {
			_ = render.New(w).TooManyRequests("Too many accounts were created at the same IP within the past 1 hour.")
			return
		}
	}

	var params input.EmailSignUpParams

	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	router.emailSignUp(w, params, clientApp)
}

// emailSignUp creates a new email account.
// Input:
// * email: string;
// * password: string;
// * mobile?: string; - Required only when mobile is linking to new account.
// * deviceToken?: string; - Required for Android app.
// * sourceUrl?: string; - Used to compose email verification link.
func (router AuthRouter) emailSignUp(w http.ResponseWriter, params input.EmailSignUpParams, client footprint.Client) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	if ve := params.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Create account from input data.
	baseAccount := account.NewEmailBaseAccount(params)
	// Save it.
	err := router.userRepo.CreateAccount(baseAccount)
	if err != nil {
		sugar.Error(err)
		// Check for duplicate.
		if db.IsAlreadyExists(err) {
			_ = render.New(w).Unprocessable(render.NewVEAlreadyExists("email"))
			return
		}
		_ = render.New(w).DBError(err)
		return
	}

	// Tracking.
	fp := footprint.New(baseAccount.FtcID, client).
		FromSignUp().
		WithAuth(enum.LoginMethodMobile, params.DeviceToken)

	go func() {
		err := router.userRepo.SaveFootprint(fp)
		if err != nil {
			sugar.Error()
		}
	}()

	// Send verification email.
	go func() {
		_ = router.SendEmailVerification(
			baseAccount,
			params.SourceURL,
			true)
	}()

	// Compose reader.Account instance.
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

	vrf, err := router.userRepo.RetrieveEmailVerifier(token)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	baseAccount, err := router.userRepo.BaseAccountByEmail(vrf.Email)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if baseAccount.IsVerified {
		_ = render.New(w).NoContent()
		return
	}

	if err := router.userRepo.EmailVerified(baseAccount.FtcID); err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	fp := footprint.
		New(baseAccount.FtcID, footprint.NewClient(req)).
		FromVerification()

	go func() {
		_ = router.userRepo.SaveFootprint(fp)
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
