package api

import (
	"net/http"

	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
)

func (routes FtcPayRoutes) LoadDiscountRedeemed(w http.ResponseWriter, req *http.Request) {
	id, _ := xhttp.GetURLParam(req, "id").ToString()

	userIDs := ids.UserIDsFromHeader(req.Header)

	redeemed, err := routes.SubsRepo.RetrieveDiscountRedeemed(
		userIDs,
		id)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(redeemed)
}
