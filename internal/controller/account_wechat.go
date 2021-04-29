package controller

import (
	"errors"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"io"
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

	var in pkg.EmailSignUpParams
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

	result, err := router.repo.WxSignUp(unionID, in)

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
			_ = router.repo.ArchiveMember(result.WxMemberSnapshot)
		}()
	}

	fp := footprint.New(result.Account.FtcID, footprint.NewClient(req)).
		FromSignUp()

	go func() {
		_ = router.repo.SaveFootprint(fp)
	}()

	// Send an email telling user that a new account is created with this email, wechat is bound to it, and in the future the email account is equal to wechat account.
	// Add customer service information so that user could contact us for any other issues.
	go func() {
		verifier, err := account.NewEmailVerifier(in.Email, in.SourceURL)
		if err != nil {
			return
		}

		err = router.repo.SaveEmailVerifier(verifier)
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

	var input pkg.LinkWxParams
	if err := gorest.ParseJSON(req.Body, &input); err != nil && err != io.EOF {
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

	result, err := router.repo.LinkWechat(input)

	if err != nil {
		var ve *render.ValidationError
		if errors.As(err, &ve) {
			_ = render.New(w).Unprocessable(ve)
			return
		}
		switch {
		case errors.As(err, &ve):
			_ = render.New(w).Unprocessable(ve)

		case err == reader.ErrAccountsAlreadyLinked:
			_ = render.New(w).NoContent()

		default:
			_ = render.New(w).DBError(err)
		}
		return
	}

	go func() {
		if !result.FtcMemberSnapshot.IsZero() {
			_ = router.repo.ArchiveMember(result.FtcMemberSnapshot)
		}

		if !result.WxMemberSnapshot.IsZero() {
			_ = router.repo.ArchiveMember(result.WxMemberSnapshot)
		}
	}()

	// Send email telling user that the accounts are linked.
	go func() {
		parcel, err := letter.LinkedParcel(result)
		if err != nil {
			sugar.Error(err)
		}

		err = router.postman.Deliver(parcel)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(result.Account)
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

	var params pkg.UnlinkWxParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
	}
	acnt, err := router.repo.AccountByFtcID(params.FtcID)
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
		sugar.Infof("Union id does not match. Expected %, got %s", params.UnionID, acnt.UnionID.String)
		_ = render.New(w).NotFound("")
		return
	}

	if ve := acnt.ValidateUnlink(params.Anchor); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	err = router.repo.UnlinkWx(acnt, params.Anchor)
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
