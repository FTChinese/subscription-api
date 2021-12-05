package controller

import (
	"errors"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"net/http"
)

// WxSignUp allows a Wechat logged in user to create a new FTC account and binds it to the Wechat account.
//
//	POST /users/wx/signup
//
// Header `X-Union-Id: <wechat union id>`
//
// Input: input.EmailSignUpParams
// * email: string;
// * password: string;
// * sourceUrl?: string;
//
// The footprint.Client headers are required.
func (router AccountRouter) WxSignUp(w http.ResponseWriter, req *http.Request) {

	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	unionID := ids.GetUnionID(req.Header)

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

	// Check if targeting email exists
	ok, err := router.Repo.EmailExists(in.Email)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if ok {
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: "Targeting email already exists",
			Field:   "email",
			Code:    render.CodeAlreadyExists,
		})
		return
	}

	// A new complete email account.
	// You should set LoginMethod to LoginMethodEmail
	// so that the Link step knows how to merge data.
	ftcAccount := reader.Account{
		BaseAccount: account.NewEmailBaseAccount(in),
		LoginMethod: enum.LoginMethodEmail,
		Wechat:      account.Wechat{},
		Membership:  reader.Membership{},
	}

	// Retrieve account by wx union id.
	wxAccount, err := router.ReaderRepo.AccountByWxID(unionID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	merged, err := ftcAccount.Link(wxAccount)
	if err != nil {
		sugar.Error(err)
		var ve *render.ValidationError
		if errors.As(err, &ve) {
			_ = render.New(w).Unprocessable(ve)
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	// Possible 422 error that could only happen for login.
	// For signup the email account is always clean so link
	// is always possible.
	// field: account_link, code: already_exists;
	// field: membership_link, code: already_exists;
	// field: membership_both_valid, code: already_exists;
	err = router.Repo.WxSignUp(merged)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if !wxAccount.Membership.IsZero() {
		go func() {
			_ = router.ReaderRepo.ArchiveMember(wxAccount.Membership.Snapshot(reader.Archiver{
				Name:   reader.ArchiveNameWechat,
				Action: reader.ArchiveActionLink,
			}))
		}()
	}

	fp := footprint.
		New(merged.FtcID, footprint.NewClient(req)).
		FromSignUp()

	go func() {
		_ = router.Repo.SaveFootprint(fp)
	}()

	// Send an email telling user that a new account is created with this email, wechat is bound to it, and in the future the email account is equal to wechat account.
	// Add customer service information so that user could contact us for any other issues.
	go func() {
		sugar.Info("Sending wechat signup email...")
		verifier, err := account.NewEmailVerifier(in.Email, in.SourceURL)
		if err != nil {
			return
		}

		err = router.Repo.SaveEmailVerifier(verifier)
		if err != nil {
			return
		}

		parcel, err := letter.WxSignUpParcel(merged, verifier)
		if err != nil {
			return
		}

		err = router.Postman.Deliver(parcel)
	}()

	// Change login method to wechat so that when unlinked, client knows which side should be used.
	merged.LoginMethod = enum.LoginMethodWx
	_ = render.New(w).OK(merged)
}

// WxLinkEmail links FTC account with a Wechat account.
//
//	POST /account/wx/link
//
// Header
// * `X-Union-Id: <wechat union id>`
//
// Input: input.LinkWxParams
// * ftcId: string;
func (router AccountRouter) WxLinkEmail(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	// Get union id from request header.
	unionID := ids.GetUnionID(req.Header)

	var params input.LinkWxParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	params.UnionID = unionID

	if ve := params.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Retrieve accounts for ftc side and wx side respectively.
	ftcAcnt, err := router.ReaderRepo.AccountByFtcID(params.FtcID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
	}

	wxAcnt, err := router.ReaderRepo.AccountByWxID(params.UnionID)
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

	err = router.Repo.LinkWechat(result)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		if !result.FtcMemberSnapshot.IsZero() {
			_ = router.ReaderRepo.ArchiveMember(result.FtcMemberSnapshot)
		}

		if !result.WxMemberSnapshot.IsZero() {
			_ = router.ReaderRepo.ArchiveMember(result.WxMemberSnapshot)
		}
	}()

	// Send email telling user that the accounts are linked.
	go func() {
		parcel, err := letter.LinkedParcel(result)
		if err != nil {
			sugar.Error(err)
		}

		sugar.Info("Sending wechat link email...")
		err = router.Postman.Deliver(parcel)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).NoContent()
}

// WxUnlinkEmail revert linking accounts.
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
func (router AccountRouter) WxUnlinkEmail(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	// Get union id from request header.
	unionID := ids.GetUnionID(req.Header)

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
	acnt, err := router.ReaderRepo.AccountByFtcID(params.FtcID)
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

	err = router.Repo.UnlinkWx(acnt, params.Anchor)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		_ = router.ReaderRepo.ArchiveMember(acnt.Membership.Snapshot(reader.Archiver{
			Name:   reader.ArchiveNameWechat,
			Action: reader.ArchiveActionUnlink,
		}))
	}()

	go func() {
		parcel, err := letter.UnlinkParcel(acnt, params.Anchor)
		if err != nil {
			sugar.Error(err)
		}

		err = router.Postman.Deliver(parcel)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).NoContent()
}
