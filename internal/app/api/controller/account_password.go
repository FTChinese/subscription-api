package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"net/http"
)

// UpdatePassword changes password after user login
//
// 	PATCH /user/password
//
// Input
// oldPassword: string; current password.
// password: string; the new password.
// * currentPassword: string;
// * newPassword: string.
func (router AccountRouter) UpdatePassword(w http.ResponseWriter, req *http.Request) {
	ftcID := ids.GetFtcID(req.Header)

	var params input.PasswordUpdateParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest("")
		return
	}

	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// 404 Not Found might occur.
	authResult, err := router.Repo.VerifyIDPassword(account.IDCredentials{
		FtcID:    ftcID,
		Password: params.CurrentPassword,
	})

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// 403 Forbidden if password incorrect.
	if !authResult.PasswordMatched {
		_ = render.New(w).Forbidden("Current password incorrect")
		return
	}

	// ErrWrongPassword if current password does not match -- 403 Forbidden.
	// 404 Not Found is user id does not exist.
	err = router.Repo.UpdatePassword(account.IDCredentials{
		FtcID:    ftcID,
		Password: params.NewPassword,
	})
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// `204 No Content`
	_ = render.New(w).NoContent()
}
