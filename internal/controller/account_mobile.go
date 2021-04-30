package controller

import (
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
	ok, err := router.repo.IDExists(ftcID)
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

	if ve := params.ValidateMobile(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	vrf := ztsms.NewVerifier(params.Mobile, null.StringFrom(ftcID))

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

// UpdateMobile set mobile_phone field to the specified number
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

	if ve := params.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	vrf, err := router.repo.RetrieveSMSVerifier(params)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if vrf.FtcID.String != ftcID {
		_ = render.New(w).NotFound("")
		return
	}

	vrf = vrf.WithUsed()
	go func() {
		err = router.repo.SMSVerifierUsed(vrf)
		sugar.Error(err)
	}()

	acnt, err := router.repo.BaseAccountByUUID(ftcID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	acnt = acnt.WithMobile(vrf.Mobile)

	err = router.repo.SetPhone(acnt)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(acnt)
}
