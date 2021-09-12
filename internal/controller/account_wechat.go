package controller

import (
	"errors"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"net/http"
)

// WxSignUp allows a Wechat logged in user to create a new FTC account and binds it to the Wechat account.
//
//	POST /users/wx/signup
//
// Header `X-Union-Id: <wechat union id>`
// Input: user.SignUpInput
// email: string;
// password: string;
// sourceUrl?: string;
//
// The footprint.Client headers are required.
func (router AccountRouter) WxSignUp(w http.ResponseWriter, req *http.Request) {

	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	unionID := req.Header.Get(unionIDKey)

	var in input.EmailSignUpParams
	// 400 Bad Request.
	if err := gorest.ParseJSON(req.Body, &in); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := in.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Possible 422 error that could only happen for login.
	// For signup the email account is always clean so link
	// is always possible.
	// field: account_link, code: already_exists;
	// field: membership_link, code: already_exists;
	// field: membership_both_valid, code: already_exists;
	result, err := router.userRepo.WxSignUp(unionID, in)

	if err != nil {
		var ve *render.ValidationError
		if errors.As(err, &ve) {
			_ = render.New(w).Unprocessable(ve)
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	if !result.WxMemberSnapshot.IsZero() {
		go func() {
			_ = router.userRepo.ArchiveMember(result.WxMemberSnapshot)
		}()
	}

	fp := footprint.New(result.Account.FtcID, footprint.NewClient(req)).
		FromSignUp()

	go func() {
		_ = router.userRepo.SaveFootprint(fp)
	}()

	// Send an email telling user that a new account is created with this email, wechat is bound to it, and in the future the email account is equal to wechat account.
	// Add customer service information so that user could contact us for any other issues.
	go func() {
		sugar.Info("Sending wechat signup email...")
		verifier, err := account.NewEmailVerifier(in.Email, in.SourceURL)
		if err != nil {
			return
		}

		err = router.userRepo.SaveEmailVerifier(verifier)
		if err != nil {
			return
		}

		parcel, err := letter.WxSignUpParcel(result.Account, verifier)
		if err != nil {
			return
		}

		err = router.postman.Deliver(parcel)

	}()

	_ = render.New(w).OK(result.Account)
}

// LinkWechat links FTC account with a Wechat account.
//
//	POST /account/wx/link
//
// Header `X-Union-Id: <wechat union id>`
//
// Input: user.LinkInput
// ```ts
// ftcId: string;
// ```
func (router AccountRouter) LinkWechat(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Get union id from request header.
	unionID := req.Header.Get(unionIDKey)

	var input input.LinkWxParams
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	input.UnionID = unionID

	if ve := input.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Retrieve accounts for ftc side and wx side respectively.
	ftcAcnt, err := router.userRepo.AccountByFtcID(input.FtcID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
	}

	wxAcnt, err := router.userRepo.AccountByWxID(input.UnionID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
	}

	// Possible 422 error that could only happen:
	// field: account_link, code: already_exists;
	// field: membership_link, code: already_exists;
	// field: membership_both_valid, code: already_exists;
	result, err := reader.WxEmailLinkBuilder{
		FTC:    ftcAcnt,
		Wechat: wxAcnt,
	}.Build()

	if err != nil {
		var ve *render.ValidationError
		switch {
		case errors.As(err, &ve):
			_ = render.New(w).Unprocessable(ve)

		default:
			_ = render.New(w).DBError(err)
		}
		return
	}

	if result.IsDuplicateLink {
		sugar.Info("Duplicate wechat-email link")
		_ = render.New(w).NoContent()
		return
	}

	err = router.userRepo.LinkWechat(result)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		if !result.FtcMemberSnapshot.IsZero() {
			_ = router.userRepo.ArchiveMember(result.FtcMemberSnapshot)
		}

		if !result.WxMemberSnapshot.IsZero() {
			_ = router.userRepo.ArchiveMember(result.WxMemberSnapshot)
		}
	}()

	// Send email telling user that the accounts are linked.
	go func() {
		parcel, err := letter.LinkedParcel(result)
		if err != nil {
			sugar.Error(err)
		}

		sugar.Info("Sending wechat link email...")
		err = router.postman.Deliver(parcel)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).NoContent()
}

// UnlinkWx revert linking accounts.
//
// PUT /user/unlink/wx
//
// Header `X-Union-Id`
//
// Request body:
// ftcId: string;
// anchor?: ftc | wechat;
//
// If user is a member and
// not expired, `anchor` field must be provided indicating
// after accounts unlinked, with which side the membership should be kept.
func (router AccountRouter) UnlinkWx(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Get union id from request header.
	unionID := req.Header.Get(unionIDKey)

	var params input.UnlinkWxParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	params.UnionID = unionID

	if ve := params.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
	}
	acnt, err := router.userRepo.AccountByFtcID(params.FtcID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// If the account is not linked, it should be treated
	// as if it is not found since in SQL if you use
	// WHERE user_id = ? AND wx_union_id = ?, the result is
	// ErrNoRow
	if !acnt.IsLinked() {
		sugar.Infof("Account not linked")
		_ = render.New(w).NotFound("")
		return
	}
	if acnt.UnionID.String != params.UnionID {
		sugar.Infof("Union id does not match. Expected %s, got %s", params.UnionID, acnt.UnionID.String)
		_ = render.New(w).NotFound("")
		return
	}

	if ve := acnt.ValidateUnlink(params.Anchor); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	err = router.userRepo.UnlinkWx(acnt, params.Anchor)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		parcel, err := letter.UnlinkParcel(acnt, params.Anchor)
		if err != nil {
			sugar.Error(err)
		}

		err = router.postman.Deliver(parcel)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).NoContent()
}
