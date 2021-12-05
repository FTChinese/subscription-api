package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"net/http"
)

// UpdateEmail lets user to change email.
//
//	PATCH /user/email
//
// Input {email: string, sourceUrl?: string}
func (router AccountRouter) UpdateEmail(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	ftcID := ids.GetFtcID(req.Header)

	var params input.EmailUpdateParams
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

	// Find current user by id.
	currAcnt, err := router.ReaderRepo.BaseAccountByUUID(ftcID)
	// AccountBase might not be found.
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if currAcnt.IsTest() {
		_ = render.New(w).Forbidden("Test account cannot change email")
		return
	}

	// If email is not actually changed
	if params.Email == currAcnt.Email {
		_ = render.New(w).NoContent()
		return
	}

	newAcnt := currAcnt.WithEmail(params.Email)
	// Update email and record email change history.
	err = router.Repo.UpdateEmail(newAcnt)

	// `422 Unprocessable Entity`
	if err != nil {
		sugar.Error(err)
		if db.IsAlreadyExists(err) {
			_ = render.New(w).Unprocessable(
				render.NewVEAlreadyExists("email"))
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	// Save user's current email address.
	go func() {
		if err := router.Repo.SaveEmailHistory(currAcnt); err != nil {
			sugar.Error(err)
		}
	}()

	go func() {
		_ = router.SendEmailVerification(newAcnt, params.SourceURL.String, false)
	}()

	// `204 No Content` if updated successfully.
	_ = render.New(w).OK(newAcnt)
}

// RequestVerification sends user a verification letter when he explicitly ask so.
// We need to tell user if email cannot be sent.
//	POST /user/email/request-verification
//
// Input {sourceUrl?: string}
func (router AccountRouter) RequestVerification(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	ftcID := ids.GetFtcID(req.Header)

	var params input.ReqEmailVrfParams

	// Ignore empty body for backward-compatibility.
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Retrieve this user info by user id.
	baseAccount, err := router.ReaderRepo.BaseAccountByUUID(ftcID)
	// 404 if user is not found.
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	err = router.SendEmailVerification(
		baseAccount,
		params.SourceURL.String,
		false)

	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	fp := footprint.New(baseAccount.FtcID, footprint.NewClient(req)).FromVerification()

	go func() {
		err = router.Repo.SaveFootprint(fp)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).NoContent()
}
