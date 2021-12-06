package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

func (router AccountRouter) LoadMembership(w http.ResponseWriter, req *http.Request) {
	userIDs := xhttp.GetUserIDs(req.Header)

	m, err := router.ReaderRepo.RetrieveMember(userIDs.CompoundID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(m)
}
