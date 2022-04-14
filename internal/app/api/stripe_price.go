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
