package api

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

func (routes StripeRoutes) ListPriceCoupons(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()
	activeOnly := xhttp.ParseQueryBool(req, "active_only")

	priceId, _ := xhttp.GetURLParam(req, "id").ToString()

	coupons, err := routes.stripeRepo.ListPriceCoupons(priceId, activeOnly)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(coupons)
}

// LoadStripeCoupon gets a stripe coupon from db or Stripe API.
// Query parameters:
// - refresh=true to force refresh db data.
func (routes StripeRoutes) LoadStripeCoupon(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	var refresh = xhttp.ParseQueryBool(req, "refresh")
	var id, _ = xhttp.GetURLParam(req, "id").ToString()

	c, err := routes.stripeRepo.LoadOrFetchCoupon(id, refresh)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	if c.IsFromStripe {
		go func() {
			err := routes.stripeRepo.UpsertCoupon(c)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(c)
}

// UpdateStripeCoupon changes a coupon's metadata.
// This will hit Stripe API before updating FTC's database.
// It adds price id, start utc and end utc to Stripe coupon's metadata.
// Do not use this to change a coupon's status field since the status field
// is not stored on Stripe's side.
// Limited only to internal use.
func (routes StripeRoutes) UpdateStripeCoupon(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	var id, _ = xhttp.GetURLParam(req, "id").ToString()

	var params price.StripeCouponMeta
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Modify the coupon against Stripe API, then upsert it in database.
	c, err := routes.stripeRepo.ModifyCoupon(id, params)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
	}

	_ = render.New(w).OK(c)
}

func (routes StripeRoutes) ActivateCoupon(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	var id, _ = xhttp.GetURLParam(req, "id").ToString()

	c, err := routes.stripeRepo.LoadOrFetchCoupon(id, false)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	c = c.Activate()

	err = routes.stripeRepo.UpdateCouponStatus(c)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
	}

	_ = render.New(w).OK(c)
}

func (routes StripeRoutes) PauseCoupon(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	var id, _ = xhttp.GetURLParam(req, "id").ToString()

	c, err := routes.stripeRepo.LoadOrFetchCoupon(id, false)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	c = c.Pause()

	err = routes.stripeRepo.UpdateCouponStatus(c)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
	}

	_ = render.New(w).OK(c)
}

// DeleteCoupon flags a stripe coupon as invalid.
// Limited only to internal usage.
func (routes StripeRoutes) DeleteCoupon(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	var id, _ = xhttp.GetURLParam(req, "id").ToString()

	c, err := routes.stripeRepo.LoadOrFetchCoupon(id, false)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	c = c.Cancelled()

	err = routes.stripeRepo.UpdateCouponStatus(c)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
	}

	_ = render.New(w).OK(c)
}
