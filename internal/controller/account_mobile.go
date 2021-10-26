package controller

import (
	"errors"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"github.com/guregu/null"
	"net/http"
)

// SMSToModifyMobile sends an SMS to the specified mobile number.
// Used to verify and link a mobile after email account logged in.
// Input:
// mobile: string;
func (router AccountRouter) SMSToModifyMobile(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	ftcID := req.Header.Get(ftcIDKey)
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
	mobileFound, err := router.userRepo.SearchByMobile(params.Mobile)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// If the mobile is found
	// It means mobile already used either by current user, or by another account.
	if mobileFound.ID.Valid {
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: "This mobile already exists",
			Field:   "mobile",
			Code:    render.CodeAlreadyExists,
		})
		return
	}

	// If the mobile was used to create an account in userinfo,
	// but it does not appear in profile table,
	// then we have to make sure that userinfo's id matches this
	// user id; otherwise we treat it as taken by another account.
	mobileFound, err = router.userRepo.SearchByEmail(account.MobileEmail(params.Mobile))
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}
	// If we found this mobile in userinfo table.
	if mobileFound.ID.Valid {
		// The mobile is already used by another id.
		if mobileFound.ID.String != ftcID {
			_ = render.New(w).Unprocessable(&render.ValidationError{
				Message: "This mobile already exists",
				Field:   "mobile",
				Code:    render.CodeAlreadyExists,
			})
			return
		}
		// No action if the two ids matches.
		// This means we allow an existing mobile user to
		// populate missing db column.
	}

	// Mobile is not found.
	// User is allowed to link current account to this mobile.
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
// When updating mobile, we must ensure this new mobile is not used by anyone else.
// If account is created from mobile directly, we should
// forbid updating mobile.
//
// after checking the SMS code sent to user's device.
// Input:
// mobile: string;
// code: string;
// deviceToken?: string.
func (router AccountRouter) UpdateMobile(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	ftcID := req.Header.Get(ftcIDKey)

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

	// Flag the verifier as used.
	go func() {
		err = router.userRepo.SMSVerifierUsed(vrf.WithUsed())
		sugar.Error(err)
	}()

	// Retrieve account for current id.
	baseAccount, err := router.userRepo.BaseAccountByUUID(ftcID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if baseAccount.IsMobileOnly() {
		_ = render.New(w).BadRequest("Mobile-created account is not allowed to switch phone number")
		return
	}

	currentMobile := baseAccount.GetMobile()
	// Already set.
	// Here's the difference from MobileLinkExistingEmail:
	// we permit overriding the account's existing mobile.
	if currentMobile == params.Mobile {
		_ = render.New(w).OK(baseAccount)
		return
	}

	err = router.userRepo.UpsertMobile(account.MobileUpdater{
		FtcID:  ftcID,
		Mobile: null.StringFrom(vrf.Mobile),
	})

	if err != nil {
		// Ensure the new mobile is not used by anyone.
		if errors.Is(err, account.ErrMobileTakenByOther) {
			_ = render.New(w).Unprocessable(&render.ValidationError{
				Message: err.Error(),
				Field:   "mobile",
				Code:    render.CodeAlreadyExists,
			})
		}
		_ = render.New(w).DBError(err)
	}

	_ = render.New(w).OK(baseAccount.WithMobile(params.Mobile))
}

// DeleteMobile sets mobile phone to NULL.
//
// Input:
// mobile: string;
func (router AccountRouter) DeleteMobile(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	ftcID := req.Header.Get(ftcIDKey)

	var params ztsms.VerifierParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// 422
	if ve := validator.New("mobile").Required().Validate(params.Mobile); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Retrieve account for current id.
	baseAccount, err := router.userRepo.BaseAccountByUUID(ftcID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if baseAccount.Mobile.String != params.Mobile {
		_ = render.New(w).NotFound("")
		return
	}

	err = router.userRepo.DeleteMobile(account.MobileUpdater{
		FtcID:  ftcID,
		Mobile: null.String{},
	})

	baseAccount.Mobile = null.String{}
	_ = render.New(w).OK(baseAccount)
}
