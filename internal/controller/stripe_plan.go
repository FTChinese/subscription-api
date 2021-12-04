package controller

import (
	"github.com/FTChinese/go-rest/render"
	"net/http"
)

// ListPrices retrieves all prices defined in Stripe.
func (router StripeRouter) ListPrices(w http.ResponseWriter, req *http.Request) {
	refresh := req.FormValue("refresh") == "true"

	prices, err := router.StripePriceRepo.ListPrices(router.Live, refresh)

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

func (router StripeRouter) LoadPrice(w http.ResponseWriter, req *http.Request) {
	refresh := req.FormValue("refresh") == "true"
	id, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	price, err := router.StripePriceRepo.LoadPrice(id, refresh)

	if err != nil {
		err := handleErrResp(w, err)
		if err == nil {
			return
		}

		_ = render.New(w).InternalServerError(err.Error())
		return
	}

	_ = render.New(w).OK(price)
}
