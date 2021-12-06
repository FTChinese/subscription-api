package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// CreateDiscount creates a new discount and update corresponding price.
func (router PaywallRouter) CreateDiscount(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	var params price.DiscountParams
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

	// Find the price for this discount first.
	ftcPrice, err := router.ReadRepo.RetrieveFtcPrice(params.PriceID, router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	discount := price.NewDiscount(params, ftcPrice.LiveMode)
	if err := router.WriteRepo.CreateDiscount(discount); err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_, err = router.WriteRepo.RefreshFtcPriceOffers(ftcPrice)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(discount)
}

func (router PaywallRouter) ListDiscounts(w http.ResponseWriter, req *http.Request) {
	priceID, err := gorest.GetQueryParam(req, "price_id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	discounts, err := router.WriteRepo.ListDiscounts(priceID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(discounts)
}

func (router PaywallRouter) DropDiscount(w http.ResponseWriter, req *http.Request) {
	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	discount, err := router.ReadRepo.LoadDiscount(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	discount = discount.Cancel()
	err = router.WriteRepo.UpdateDiscount(discount)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	ftcPrice, err := router.ReadRepo.RetrieveFtcPrice(discount.PriceID, router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	ftcPrice, err = router.WriteRepo.RefreshFtcPriceOffers(ftcPrice)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(ftcPrice)
}
