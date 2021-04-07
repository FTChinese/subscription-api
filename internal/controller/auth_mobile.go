package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/account"
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

	_, err = router.client.SendVerifier(vrf)
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
