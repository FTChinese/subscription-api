package api

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// CreateDiscount creates a new discount and update corresponding price.
func (router PaywallRouter) CreateDiscount(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

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
	ftcPrice, err := router.paywallRepo.RetrievePaywallPrice(params.PriceID, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	discount := price.NewDiscount(params, ftcPrice.LiveMode)
	if err := router.productRepo.CreateDiscount(discount); err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_, err = router.productRepo.RefreshPriceOffers(ftcPrice)
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

	discounts, err := router.productRepo.
		ListDiscounts(priceID, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(discounts)
}

func (router PaywallRouter) LoadDiscount(w http.ResponseWriter, req *http.Request) {
	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	discount, err := router.paywallRepo.LoadDiscount(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(discount)
}

func (router PaywallRouter) DropDiscount(w http.ResponseWriter, req *http.Request) {
	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	discount, err := router.paywallRepo.LoadDiscount(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	discount = discount.Cancel()
	err = router.productRepo.UpdateDiscount(discount)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	ftcPrice, err := router.paywallRepo.RetrievePaywallPrice(discount.PriceID, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	ftcPrice, err = router.productRepo.RefreshPriceOffers(ftcPrice)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(ftcPrice)
}
