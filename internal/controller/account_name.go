package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/db"
	"net/http"
)

// UpdateName handles user changing name
//
// 	PATH /user/name
//
// Input
// userName: string
func (router AccountRouter) UpdateName(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(userIDKey)

	var params pkg.NameUpdateParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	baseAccount, err := router.userRepo.BaseAccountByUUID(ftcID)
	if err != nil {
		_ = render.New(w).NotFound("")
		return
	}

	if baseAccount.UserName.String == params.UserName {
		_ = render.New(w).NoContent()
		return
	}

	baseAccount = baseAccount.WithUserName(params.UserName)
	err = router.userRepo.UpdateUserName(baseAccount)

	// `422 Unprocessable Entity` if this `userName` already exists
	if err != nil {
		if db.IsAlreadyExists(err) {
			_ = render.New(w).Unprocessable(render.NewVEAlreadyExists("userName"))
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	// `204 No Content`
	_ = render.New(w).OK(baseAccount)
}
