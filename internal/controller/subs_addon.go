package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"net/http"
)

func (router SubsRouter) ClaimAddOn(w http.ResponseWriter, req *http.Request) {
	readerIDs := getReaderIDs(req.Header)

	result, err := router.AddOnRepo.ClaimAddOn(readerIDs)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		_ = router.SubsRepo.ArchiveMember(result.Snapshot)
	}()

	_ = render.New(w).OK(result.Membership)
}

// CreateAddOn manually add an addon to a user.
// This is usually used to perform compensation.
func (router SubsRouter) CreateAddOn(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	readerIDs := getReaderIDs(req.Header)

	var params invoice.AddOnParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}
	params.CompoundID = readerIDs.CompoundID

	inv := invoice.NewAddonInvoice(params)

	result, err := router.AddOnRepo.CreateAddOn(inv)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		err := router.SubsRepo.ArchiveMember(result.Snapshot)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(result)
}
