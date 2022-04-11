package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// ListPrices retrieves all prices defined in Stripe.
// Deprecated
func (router StripeRouter) ListPrices(w http.ResponseWriter, req *http.Request) {
	refresh := xhttp.ParseQueryRefresh(req)

	prices, err := router.stripeRepo.ListPricesCompat(router.live, refresh)

	if err != nil {
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	if len(prices) == 0 {
		_ = render.New(w).NotFound("No prices found")
		return
	}

	_ = render.New(w).OK(prices)
}

// LoadPrice tries to find a single price of Stripe.
func (router StripeRouter) LoadPrice(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	refresh := req.FormValue("refresh") == "true"
	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	price, err := router.stripeRepo.LoadCachedPrice(id, refresh)

	if err != nil {
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	if price.IsFromStripe {
		go func() {
			err := router.stripeRepo.UpsertPrice(price)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(price)
}
