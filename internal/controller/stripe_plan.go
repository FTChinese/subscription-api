package controller

import (
	"github.com/FTChinese/go-rest/render"
	"net/http"
)

// ListPrices retrieves all prices defined in Stripe.
func (router StripeRouter) ListPrices(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	refresh := req.FormValue("refresh") == "true"

	prices, err := router.stripeRepo.ListPrices(refresh)

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
