package api

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

func (routes StripeRoutes) ListCouponsOfPrice(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	priceId, _ := xhttp.GetURLParam(req, "id").ToString()

	coupons, err := routes.stripeRepo.RetrieveCouponsOfPrice(priceId)
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

	c, err := routes.stripeRepo.ModifyCoupon(id, params)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
	}

	_ = render.New(w).OK(c)
}

// DeleteStripeCoupon flags a stripe coupon as invalid.
// Limited only to internal usagte.
func (routes StripeRoutes) DeleteStripeCoupon(w http.ResponseWriter, req *http.Request) {
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

	err = routes.stripeRepo.UpsertCoupon(c)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
	}

	_ = render.New(w).OK(c)
}
