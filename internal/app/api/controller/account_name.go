package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// UpdateName handles user changing name
//
// 	PATH /user/name
//
// Input
// userName: string
func (router AccountRouter) UpdateName(w http.ResponseWriter, req *http.Request) {
	ftcID := xhttp.GetFtcID(req.Header)

	var params input.NameUpdateParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	baseAccount, err := router.ReaderRepo.BaseAccountByUUID(ftcID)
	if err != nil {
		_ = render.New(w).NotFound("")
		return
	}

	if baseAccount.UserName.String == params.UserName {
		_ = render.New(w).NoContent()
		return
	}

	baseAccount = baseAccount.WithUserName(params.UserName)
	err = router.Repo.UpdateUserName(baseAccount)

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
