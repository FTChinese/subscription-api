package api

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// ForgotPassword checks user's email and send a password
// reset letter if it is valid.
//
//	POST /users/password-reset/letter
//
// Input:
// * email: string;
// * useCode: boolean; - A short number string used on native apps to facilitate input.
// * sourceUrl?: string; - Only applicable to web app.
//
// The footprint.Client headers are required.
func (router AuthRouter) ForgotPassword(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	var params input.ForgotPasswordParams

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

	// A fallback resort to handle a strange problem in
	// android app being unable to stringify JSON userCode field.
	client := footprint.NewClient(req)
	if client.IsApp() && !params.UseCode {
		params.UseCode = true
	}

	// Load account for this email
	baseAccount, err := router.Repo.BaseAccountByEmail(params.Email)
	// Not Found.
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Generate token
	session, err := account.NewPwResetSession(params)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Save password reset  token.
	if err := router.Repo.SavePwResetSession(session); err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	fp := footprint.New(baseAccount.FtcID, client).
		FromPwReset()

	go func() {
		_ = router.Repo.SaveFootprint(fp)
	}()

	// Compose email
	err = router.EmailService.SendPasswordReset(baseAccount, session)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// `204 No Content`
	_ = render.New(w).NoContent()
}

// VerifyResetToken verifies a password reset link.
//
// 	GET /auth/password-reset/tokens/{token}
func (router AuthRouter) VerifyResetToken(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	token, err := xhttp.GetURLParam(req, "token").ToString()
	// `400 Bad Request`
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	session, err := router.Repo.PwResetSessionByToken(token)
	// `404 Not Found`
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if session.IsUsed || session.IsExpired() {
		_ = render.NewNotFound("token already used or expired")
		return
	}

	_ = render.New(w).OK(session)
}

// VerifyResetCode verified verification code to allow
// resetting password.
//
// GET /users/password-reset/codes?email=xxx&code=xxx
func (router AuthRouter) VerifyResetCode(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	if err := req.ParseForm(); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	var params input.AppResetPwSessionParams
	if err := decoder.Decode(&params, req.Form); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	sugar.Infof("Input %v", params)

	if ve := params.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	session, err := router.Repo.PwResetSessionByCode(params)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if session.IsUsed || session.IsExpired() {
		sugar.Info("Password reset code already used or expired")
		_ = render.NewNotFound("code already used or expired")
		return
	}

	// Send token to client so that it send the token back
	// together with the new password.
	// In this way we could keep it backward-compatible
	// when calling ResetPassword.
	_ = render.New(w).OK(session)
}

// ResetPassword allow user to change password.
//
//	POST /users/password-reset
//
// Input
// * token: string;
// * password: string.
func (router AuthRouter) ResetPassword(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	var params input.PasswordResetParams

	// `400 Bad Request`
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

	// Should we check expiration time now?
	session, err := router.Repo.PwResetSessionByToken(params.Token)
	// Not found error.
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Find reader's account by the email stored under the token.
	baseAccount, err := router.Repo.BaseAccountByEmail(session.Email)
	// Might not be found.
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// Change password.
	err = router.Repo.UpdatePassword(account.IDCredentials{
		FtcID:    baseAccount.FtcID,
		Password: params.Password,
	})
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// Invalidate token.
	go func() {
		_ = router.Repo.DisablePasswordReset(params.Token)
	}()

	// `204 No Content`
	_ = render.New(w).NoContent()
}
