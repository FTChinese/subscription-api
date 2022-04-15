package api

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

func (routes FtcPayRoutes) ListInvoices(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	p := gorest.GetPagination(req)
	userIDs := xhttp.UserIDsFromHeader(req.Header)

	list, err := routes.AddOnRepo.ListInvoices(
		userIDs,
		p,
	)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
}

func (routes FtcPayRoutes) LoadInvoice(w http.ResponseWriter, req *http.Request) {
	userIDs := xhttp.UserIDsFromHeader(req.Header)

	invID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	inv, err := routes.AddOnRepo.LoadInvoice(invID)
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
