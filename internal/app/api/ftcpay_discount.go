package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

func (routes FtcPayRoutes) LoadDiscountRedeemed(w http.ResponseWriter, req *http.Request) {
	id, _ := xhttp.GetURLParam(req, "id").ToString()

	userIDs := xhttp.UserIDsFromHeader(req.Header)

	redeemed, err := routes.SubsRepo.RetrieveDiscountRedeemed(
		userIDs,
		id)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(redeemed)
}
