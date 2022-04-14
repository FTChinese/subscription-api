package api

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// LoadStripePrice retrieves a stripe price either from database or
// stripe API.
// Query parameters:
// - refresh=true
func (router CMSRouter) LoadStripePrice(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	refresh := req.FormValue("refresh") == "true"
	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	sp, err := router.stripeRepo.LoadOrFetchPrice(id, refresh)

	if err != nil {
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	if sp.IsFromStripe {
		go func() {
			err := router.stripeRepo.UpsertPrice(sp)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(sp)
}

func (router CMSRouter) LoadStripeCoupon(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var refresh = xhttp.ParseQueryBool(req, "refresh")
	var id, _ = xhttp.GetURLParam(req, "id").ToString()

	c, err := router.stripeRepo.LoadOrFetchCoupon(id, refresh)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	if c.IsFromStripe {
		go func() {
			err := router.stripeRepo.UpsertCoupon(c)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(c)
}

func (router CMSRouter) UpdateStripeCoupon(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var id, _ = xhttp.GetURLParam(req, "id").ToString()

	var params price.StripeCouponMeta
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	c, err := router.stripeRepo.ModifyCoupon(id, params)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
	}

	_ = render.New(w).OK(c)
}

func (router CMSRouter) DeleteStripeCoupon(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var id, _ = xhttp.GetURLParam(req, "id").ToString()

	c, err := router.stripeRepo.LoadOrFetchCoupon(id, false)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	c = c.Cancelled()

	err = router.stripeRepo.UpsertCoupon(c)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
	}

	_ = render.New(w).OK(c)
}
