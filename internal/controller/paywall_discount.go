package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
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
	ftcPrice, err := router.repo.RetrieveFtcPrice(params.PriceID, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	discount := price.NewDiscount(params, ftcPrice.LiveMode)
	if err := router.repo.CreateDiscount(discount); err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_, err = router.repo.RefreshFtcPriceOffers(ftcPrice)
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

	discounts, err := router.repo.ListDiscounts(priceID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(discounts)
}

// DropDiscount flags a discount as inactive and refresh all discounts saved
// as JSON under its parent price.
func (router PaywallRouter) DropDiscount(w http.ResponseWriter, req *http.Request) {
	// Get discount id
	id, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Retrieve discount
	discount, err := router.repo.LoadDiscount(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// Flag the discount as cancelled.
	discount = discount.Cancel()
	err = router.repo.UpdateDiscount(discount)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// Use the discount's price id to retrieve its parent price.
	ftcPrice, err := router.repo.RetrieveFtcPrice(discount.PriceID, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// Refresh the discount_list column.
	ftcPrice, err = router.repo.RefreshFtcPriceOffers(ftcPrice)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// Return the updated price together with discount list.
	_ = render.New(w).OK(ftcPrice)
}
