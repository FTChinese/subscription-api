package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"github.com/guregu/null"
	"net/http"
)

// RequestSMSVerification sends a SMS to user.
// Input:
// mobile: string
func (router AuthRouter) RequestSMSVerification(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var params ztsms.VerifierParams

	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.ValidateMobile(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Retrieve account by mobile number.
	// If not found, it indicates the mobile is used for the first time.
	ftcAccount, err := router.repo.BaseAccountByMobile(params.Mobile)
	if err != nil {
		sugar.Error(err)
	}

	vrf := ztsms.NewVerifier(params.Mobile, null.NewString(ftcAccount.FtcID, ftcAccount.FtcID != ""))

	err = router.repo.SaveSMSVerifier(vrf)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	_, err = router.smsClient.SendVerifier(vrf)
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	_ = render.New(w).NoContent()
}

// VerifySMSCode verifies a code sent to user mobile devices.
// Input:
// mobile: string
// code: string
// deviceToken: string - only required for Android devices.
func (router AuthRouter) VerifySMSCode(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var params ztsms.VerifierParams

	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	vrf, err := router.repo.RetrieveSMSVerifier(params)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if !vrf.Valid() {
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: "Verification code expired",
			Field:   "code",
			Code:    render.CodeInvalid,
		})
		return
	}

	go func() {
		err := router.repo.SMSVerifierUsed(vrf.WithUsed())
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(account.NewSearchResult(vrf.FtcID.String))
}

// LinkMobile authenticates an existing email account, and link to
// the mobile phone which is used to login for the first time.
// Input:
// email: string;
// password: string;
// mobile: string;
// deviceToken?: string;
// sourceUrl?: string; // Used to compose email verification link.
func (router AuthRouter) LinkMobile(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var params pkg.MobileLinkParams
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

	authResult, err := router.repo.Authenticate(params.EmailLoginParams)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Use existing account
	if !authResult.PasswordMatched {
		_ = render.New(w).Forbidden("Incorrect credentials")
		return
	}

	acnt, err := router.repo.AccountByFtcID(authResult.UserID)
	if err != nil {
		// There shouldn't be ErrNoRow error here.
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Update account field in application.
	baseAccount := acnt.WithMobile(params.Mobile)
	acnt.BaseAccount = baseAccount

	// Save mobile number
	go func() {
		err := router.repo.SetPhone(baseAccount)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(acnt)
}

// MobileSignUp creates a new email account.
func (router AuthRouter) MobileSignUp(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var params pkg.MobileSignUpParams
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

	// Create account from input data.
	baseAccount := account.NewMobileBaseAccount(params)
	// Save it.
	err := router.repo.CreateAccount(baseAccount)
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
	clientApp := footprint.NewClient(req)
	fp := footprint.New(baseAccount.FtcID, clientApp).
		FromSignUp().
		WithAuth(enum.LoginMethodMobile, params.DeviceToken)

	go func() {
		err := router.repo.SaveFootprint(fp)
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

	// Compose an reader.Account instance.
	_ = render.New(w).OK(reader.Account{
		BaseAccount: baseAccount,
		LoginMethod: enum.LoginMethodEmail,
		Wechat:      account.Wechat{},
		Membership:  reader.Membership{},
	})
}
