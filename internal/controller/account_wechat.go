package controller

import (
	"errors"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/client"
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

	footprint := account.ClientFootprint{
		FtcID:      result.Account.FtcID,
		Client:     client.NewClientApp(req),
		CreatedUTC: chrono.TimeNow(),
		Source:     account.FootprintSourceSignUp,
	}

	go func() {
		_ = router.repo.SaveClient(footprint)
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
//	PUT /wx/link
//
// Header `X-Union-Id: <wechat union id>`
//
// Input: user.LinkInput
// ```ts
// userId: string;
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

	_ = render.New(w).NoContent()
}

func getUnlinkInput(req *http.Request) (pkg.UnlinkWxParams, error) {
	// TODO: Will be removed by the end of 2020.
	ftcID := req.Header.Get(userIDKey)
	// Only for backward compatibility. It is moved to request body.
	unionID := req.Header.Get(unionIDKey)

	var input pkg.UnlinkWxParams
	if err := gorest.ParseJSON(req.Body, &input); err != nil && err != io.EOF {
		return input, err
	}
	// Union id is always acquired from header.
	input.UnionID = unionID
	// Backward compatibility.
	// Previously ftc id is set on header. It is moved to request body later to avoid setting too much data in header.
	if input.FtcID == "" && ftcID != "" {
		input.FtcID = ftcID
	}

	return input, nil
}

// UnlinkWx revert linking accounts.
//
// PUT /user/unlink/wx
//
// Header `X-Union-Id`
//
// Request body:
// unionId: string;
// anchor?: ftc | wechat;
//
// If user is a member and
// not expired, `anchor` field must be provided indicating
// after accounts unlinked, with which side the membership should be kept.
func (router AccountRouter) UnlinkWx(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	input, err := getUnlinkInput(req)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := input.Validate(); ve != nil {
		sugar.Error(err)
		_ = render.New(w).Unprocessable(ve)
	}
	account, err := router.repo.AccountByFtcID(input.FtcID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// If the account is not linked, it should be treated
	// as if it is not found since in SQL if you use
	// WHERE user_id = ? AND wx_union_id = ?, the result is
	// ErrNoRow
	if !account.IsLinked() {
		sugar.Infof("Account not linked")
		_ = render.New(w).NotFound("")
		return
	}
	if account.UnionID.String != input.UnionID {
		sugar.Infof("Union id does not match. Expected %, got %s", input.UnionID, account.UnionID.String)
		_ = render.New(w).NotFound("")
		return
	}

	if ve := account.ValidateUnlink(input.Anchor); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	err = router.repo.UnlinkWx(account, input.Anchor)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		parcel, err := letter.UnlinkParcel(account, input.Anchor)
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
