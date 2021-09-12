package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"net/http"
)

// UpdatePassword changes password after user login
//
// 	PATCH /user/password
//
// Input
// oldPassword: string; current password.
// password: string; the new password.
func (router AccountRouter) UpdatePassword(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get(ftcIDKey)

	var input input.PasswordUpdateParams
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		_ = render.New(w).BadRequest("")
		return
	}

	if ve := input.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	input.FtcID = userID

	// 404 Not Found might occur.
	authResult, err := router.userRepo.VerifyPassword(input)
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
	if err := router.userRepo.UpdatePassword(input); err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// `204 No Content`
	_ = render.New(w).NoContent()
}
