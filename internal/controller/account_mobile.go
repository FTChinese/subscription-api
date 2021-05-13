package controller

import (
	"database/sql"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"github.com/guregu/null"
	"net/http"
)

// RequestSMSVerification sends a SMS to the specified mobile number.
// Input:
// mobile: string;
func (router AccountRouter) RequestSMSVerification(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	ftcID := req.Header.Get(userIDKey)
	ok, err := router.userRepo.IDExists(ftcID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if !ok {
		sugar.Error(err)
		_ = render.New(w).NotFound("Account not found")
		return
	}

	var params ztsms.VerifierParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// 422
	if ve := params.ValidateMobile(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Ensure the mobile is not used by an account yet.
	_, err = router.userRepo.BaseAccountByMobile(params.Mobile)
	if err == nil {
		// Account is retrieve by mobile. It means mobile already used either by current user, or by another account.
		// 422
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: "This mobile already exists",
			Field:   "mobile",
			Code:    render.CodeAlreadyExists,
		})
		return
	}
	if err != sql.ErrNoRows {
		_ = render.New(w).DBError(err)
		return
	}

	vrf := ztsms.NewVerifier(params.Mobile, null.StringFrom(ftcID))

	err = router.userRepo.SaveSMSVerifier(vrf)
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

// UpdateMobile set mobile_phone field to the specified number.
// When updating mobile, we must ensure this mobile is not used by anyone else.
//
// after checking the SMS code sent to user's device.
// Input:
// mobile: string;
// code: string;
// deviceToken?: string.
func (router AccountRouter) UpdateMobile(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	ftcID := req.Header.Get(userIDKey)

	var params ztsms.VerifierParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// 422
	if ve := params.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	vrf, err := router.userRepo.RetrieveSMSVerifier(params)
	// 404 verification code not found.
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// If the verifier is not targeting this user.
	// 404 verification code
	if vrf.FtcID.String != ftcID {
		_ = render.New(w).NotFound("")
		return
	}

	// Ensure the mobile is not set to any other account.
	// What we want is that this mobile does not exist.
	mobileAccount, err := router.userRepo.BaseAccountByMobile(vrf.Mobile)
	// An account exist under this mobile
	if err == nil {
		// Let's see if the mobile account is this user.
		if mobileAccount.FtcID == ftcID {
			sugar.Info("Mobile already set on this user")
			// User already have mobile set.
			_ = render.New(w).OK(mobileAccount)
			return
		}

		// Mobile is set on another account
		sugar.Info("Mobile is used by another account")
		// 422
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: "This mobile is already used by another accmount",
			Field:   "mobile",
			Code:    render.CodeAlreadyExists,
		})
		return
	}

	// ErrNoRows is what we want.
	if err != sql.ErrNoRows {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Flag the verifier as used.
	vrf = vrf.WithUsed()
	go func() {
		err = router.userRepo.SMSVerifierUsed(vrf)
		sugar.Error(err)
	}()

	// Retrieve account for current id.
	baseAccount, err := router.userRepo.BaseAccountByUUID(ftcID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// If this account have mobile set.
	if baseAccount.Mobile.Valid {
		// The current mobile matches verifier's mobile.
		// Return immediately so that we won't wast resources.
		if baseAccount.Mobile.String == vrf.Mobile {
			_ = render.New(w).OK(baseAccount)
			return
		}
		// Otherwise use is changing mobile.
	}

	baseAccount = baseAccount.WithMobile(vrf.Mobile)

	err = router.userRepo.SetPhone(baseAccount)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(baseAccount)
}
