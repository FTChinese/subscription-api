package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

func (routes FtcPayRoutes) IsDiscountRedeemed(w http.ResponseWriter, req *http.Request) {
	id, _ := xhttp.GetURLParam(req, "id").ToString()

	userIDs := xhttp.UserIDsFromHeader(req.Header)

	ok, err := routes.SubsRepo.IsDiscountRedeemed(userIDs, id)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(map[string]bool{
		"redeemed": ok,
	})
}
