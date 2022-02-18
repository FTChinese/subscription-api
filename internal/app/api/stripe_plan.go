package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// ListPrices retrieves all prices defined in Stripe.
func (router StripeRouter) ListPrices(w http.ResponseWriter, req *http.Request) {
	refresh := xhttp.ParseQueryRefresh(req)

	prices, err := router.Env.ListPrices(router.Live, refresh)

	if err != nil {
		_ = xhttp.HandleStripeErr(w, err)
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
	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	price, err := router.Env.LoadPrice(id, refresh)

	if err != nil {
		_ = xhttp.HandleStripeErr(w, err)
		return
	}

	_ = render.New(w).OK(price)
}
