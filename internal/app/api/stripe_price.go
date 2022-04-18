package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// ListPrices retrieves all prices defined in Stripe.
// Deprecated
func (routes StripeRoutes) ListPrices(w http.ResponseWriter, req *http.Request) {
	refresh := xhttp.ParseQueryRefresh(req)

	prices, err := routes.stripeRepo.ListPricesCompat(routes.live, refresh)

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

// LoadStripePrice retrieves a stripe price either from database or
// stripe API.
// Query parameters:
// - refresh=true
func (routes StripeRoutes) LoadStripePrice(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	refresh := req.FormValue("refresh") == "true"
	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	sp, err := routes.stripeRepo.LoadOrFetchPrice(id, refresh)

	if err != nil {
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	if sp.IsFromStripe {
		go func() {
			err := routes.stripeRepo.UpsertPrice(sp)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(sp)
}
