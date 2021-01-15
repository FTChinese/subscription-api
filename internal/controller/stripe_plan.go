package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"net/http"
)

func (router StripeRouter) ListPrices(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	refresh := req.FormValue("refresh") == "true"

	var prices []stripe.Price
	var err error
	if refresh {
		prices, err = router.stripeRepo.RefreshPrices()
	} else {
		prices, err = router.stripeRepo.ListPrices()
	}

	if err != nil {
		err := handleErrResp(w, err)
		if err == nil {
			return
		}

		_ = render.New(w).InternalServerError(err.Error())
		return
	}

	if len(prices) == 0 {
		_ = render.New(w).NotFound("No prices found")
		return
	}

	_ = render.New(w).OK(prices)
}
