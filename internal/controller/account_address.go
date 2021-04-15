package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/account"
	"net/http"
)

// LoadAddress get a user's address.
//
//	GET /user/address
func (router AccountRouter) LoadAddress(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(userIDKey)

	addr, err := router.repo.LoadAddress(ftcID)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(addr)
}

// UpdateAddress lets user to change where he/she lives
//
//	PATCH /user/address
//
// Input
// country?: string;
// province?: string;
// city?: string;
// district?: string;
// street?: string;
// postcode?: string
func (router AccountRouter) UpdateAddress(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(userIDKey)

	var addr account.Address

	if err := gorest.ParseJSON(req.Body, &addr); err != nil {
		_ = render.New(w).BadRequest(err.Error())

		return
	}

	if ve := addr.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}
	addr.FtcID = ftcID

	if err := router.repo.UpdateAddress(addr); err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// `204 No Content`
	_ = render.New(w).OK(addr)
}
