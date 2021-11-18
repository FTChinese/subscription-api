package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"net/http"
)

// ListPrices retrieves all prices under a product.
func (router PaywallRouter) ListPrices(w http.ResponseWriter, req *http.Request) {
	productID, err := gorest.GetQueryParam(req, "product_id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	prices, err := router.repo.ListPrices(productID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(prices)
}

// CreatePrice creates a new price.
func (router PaywallRouter) CreatePrice(w http.ResponseWriter, req *http.Request) {
	var params price.FtcPriceParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Attach the strip price id to ftc price.
	se, err := stripe.PriceEditionStore.FindByEdition(params.Edition, params.LiveMode)
	if err != nil {
		_ = render.NewInternalError(err.Error())
		return
	}

	p := price.NewFtcPrice(params, se.PriceID)

	err = router.repo.CreatePrice(p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(p)
}

// RefreshPrice discounts and stripe price id.
func (router PaywallRouter) RefreshPrice(w http.ResponseWriter, req *http.Request) {
	priceID, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Find the price for this discount first.
	ftcPrice, err := router.repo.RetrieveFtcPrice(priceID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// Update offers
	ftcPrice, err = router.repo.RefreshFtcPriceOffers(ftcPrice)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	se, err := stripe.PriceEditionStore.
		FindByEdition(
			ftcPrice.Edition,
			ftcPrice.LiveMode,
		)

	// Update stripe price id.
	if err == nil {
		ftcPrice = ftcPrice.WithStripePrice(se.PriceID)
		_ = router.repo.UpdateFtcPrice(ftcPrice)
	}

	_ = render.New(w).OK(ftcPrice)
}

func (router PaywallRouter) ActivatePrice(w http.ResponseWriter, req *http.Request) {
	priceID, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	ftcPrice, err := router.repo.ActivatePrice(priceID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(ftcPrice)
}

func (router PaywallRouter) ArchivePrice(w http.ResponseWriter, req *http.Request) {
	priceID, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	p, err := router.repo.RetrieveFtcPrice(priceID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	p = p.Archive()

	err = router.repo.ArchivePrice(p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	err = router.repo.ArchivePriceDiscounts(p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(p)
}
