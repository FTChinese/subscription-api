package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/account"
	"net/http"
)

// LoadProfile shows a user's profile.
// Request header must contain `X-User-Id`.
//
//	 GET `/user/profile`
func (router AccountRouter) LoadProfile(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get(userIDKey)

	p, err := router.repo.LoadProfile(userID)

	// `404 Not Found` if this user does not exist.
	if err != nil {
		_ = render.New(w).DBError(err)

		return
	}

	_ = render.New(w).OK(p)
}

// UpdateProfile update a user's profile.
//
//	PATCH /user/profile
//
// Input:
// familyName?: string;
// givenName?: string;
// birthday?: string;
// gender?: M | F;
func (router AccountRouter) UpdateProfile(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(userIDKey)

	var input account.BaseProfile
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		_ = render.New(w).BadRequest("")
		return
	}

	if r := input.Validate(); r != nil {
		_ = render.New(w).Unprocessable(r)
		return
	}

	input.FtcID = ftcID

	err := router.repo.UpdateProfile(input)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// 204 Not Content
	_ = render.New(w).OK(input)
}
