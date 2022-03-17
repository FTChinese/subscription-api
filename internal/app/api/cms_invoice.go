package api

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"net/http"
)

// CreateAddOn manually add an addon to a user.
// This is usually used to perform compensation.
func (router CMSRouter) CreateAddOn(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	var params invoice.AddOnParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	inv := invoice.NewAddonInvoice(params)

	result, err := router.Repo.CreateAddOn(inv)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		err := router.ReaderRepo.VersionMembership(result.Versioned)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(result)
}
