package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"net/http"
)

func (router SubsRouter) LoadMembership(w http.ResponseWriter, req *http.Request) {
	userIDs := getReaderIDs(req.Header)

	m, err := router.SubsRepo.RetrieveMember(userIDs.CompoundID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(m)
}

func (router SubsRouter) UpdateMembership(w http.ResponseWriter, req *http.Request) {
	_ = render.New(w).BadRequest("Not implemented")
}

func (router SubsRouter) CreateMembership(w http.ResponseWriter, req *http.Request) {
	_ = render.New(w).BadRequest("Not implemented")
}

// ListMemberSnapshots loads a list of membership change history.
// Pagination support by adding query parameter:
// page=<int>&per_page=<int>
func (router SubsRouter) ListMemberSnapshots(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	p := gorest.GetPagination(req)
	userIDs := getReaderIDs(req.Header)

	list, err := router.SubsRepo.ListSnapshot(userIDs, p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
}
