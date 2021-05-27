package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"net/http"
)

func (router SubsRouter) ListInvoices(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	p := gorest.GetPagination(req)
	userIDs := getReaderIDs(req.Header)

	list, err := router.AddOnRepo.ListInvoices(
		userIDs,
		p,
	)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
}

func (router SubsRouter) CreateInvoice(w http.ResponseWriter, req *http.Request) {
	_ = render.New(w).BadRequest("Not implemented")
}

func (router SubsRouter) LoadInvoice(w http.ResponseWriter, req *http.Request) {
	userIDs := getReaderIDs(req.Header)

	invID, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	inv, err := router.AddOnRepo.LoadInvoice(invID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if inv.CompoundID != userIDs.CompoundID {
		_ = render.New(w).NotFound("")
		return
	}

	_ = render.New(w).OK(inv)
}
