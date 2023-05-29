package api

import (
	"net/http"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
)

// ListPaywallPrices retrieves all prices defined in Stripe.
// Deprecated
func (routes StripeRoutes) ListPaywallPrices(w http.ResponseWriter, req *http.Request) {
	refresh := xhttp.ParseQueryRefresh(req)

	prices, err := routes.stripeRepo.LoadOrFetchPaywallPrices(refresh, routes.live)

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

// ListPricesPages retrieves a list of stripe prices with pagination.
// Query parameter:
// ?page=<int>&per_page=<int>
func (routes StripeRoutes) ListPricesPaged(w http.ResponseWriter, req *http.Request) {
	p := gorest.GetPagination(req)

	prices, err := routes.stripeRepo.ListPricesPaged(routes.live, p)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(prices)
}

// LoadPrice retrieves a stripe price either from database or
// stripe API.
// Query parameters:
// - refresh=true
func (routes StripeRoutes) LoadPrice(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	refresh := xhttp.ParseQueryRefresh(req)

	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	sp, err := routes.stripeRepo.LoadOrFetchPrice(id, refresh, routes.live)

	if err != nil {
		return
	}

	_ = render.New(w).OK(sp)
}

// Set a stripe price's metadata.
//
// For introductory price, you have to provide:
// - introductory: true
// - tier: string
// - periodCount.years: int
// - periodCount.months: int
// - periodCount.days: int
// - startUtc: string
// - endUtc: string
//
// For recurring price, you only provide:
// - introductory: false
// - tier: string
// Since periodCount fields could be deduced from
// stripe price fields in such case,
// We'd better not touch it to avoid any data inconsistency.
func (routes StripeRoutes) SetPriceMeta(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	// Get body and validate it.
	var params price.StripePriceMeta
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	sp, err := routes.stripeRepo.LoadOrFetchPrice(id, false, routes.live)
	if err != nil {
		return
	}

	// Use client period count only for introductory.
	if !params.Introductory {
		params.PeriodCount = sp.PeriodCount.YearMonthDay
	}

	rawPrice, err := routes.stripeRepo.Client.SetPriceMeta(id, params.ToParams())
	if err != nil {
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	// Refresh db with updated data.
	newPrice := price.NewStripePrice(rawPrice)
	go func() {
		err := routes.stripeRepo.UpsertPrice(sp)
		if err != nil {
			defer routes.logger.Sync()
			sugar := routes.logger.Sugar()
			sugar.Error(err)
		}
	}()

	newPrice.OnPaywall = sp.OnPaywall

	_ = render.New(w).OK(newPrice)
}

func (routes StripeRoutes) ActivatePrice(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	sp, err := routes.stripeRepo.LoadOrFetchPrice(id, false, routes.live)

	if err != nil {
		return
	}

	// Put this price on paywall.
	err = routes.stripeRepo.ActivatePrice(sp)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(sp)
}

func (routes StripeRoutes) DeactivatePrice(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	sp, err := routes.stripeRepo.LoadOrFetchPrice(id, false, routes.live)

	if err != nil {
		return
	}

	// Remove price from paywall.
	err = routes.stripeRepo.DeactivePrice(sp)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(sp)
}
