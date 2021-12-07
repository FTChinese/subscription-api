package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// ListPrices retrieves all prices under a product.
func (router PaywallRouter) ListPrices(w http.ResponseWriter, req *http.Request) {
	productID, err := gorest.GetQueryParam(req, "product_id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	prices, err := router.ProductRepo.ListPrices(productID, router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(prices)
}

// CreatePrice creates a new price.
// Request body:
// - createdBy: string
// - tier: standard | premium
// - cycle: year | month
// - description?: string
// - liveMode: boolean
// - nickname?: string
// - productId: string
// - stripePriceId: string
// - unitAmount: number
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

	p := price.NewFtcPrice(params, router.Live)

	err := router.ProductRepo.CreatePrice(p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(p)
}

// UpdatePrice changes a price's Description, Nickname, or StripePriceID fields
// Input body:
// - description?: string;
// - nickname?: string;
// - stripePriceId: string;
func (router PaywallRouter) UpdatePrice(w http.ResponseWriter, req *http.Request) {
	id, _ := xhttp.GetURLParam(req, "id").ToString()

	var params price.FtcPriceUpdateParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	ftcPrice, err := router.PaywallRepo.RetrieveFtcPrice(id, router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	updated := ftcPrice.Update(params)
	err = router.ProductRepo.UpdateFtcPrice(updated)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(updated)
}

// RefreshPriceOffers attaches all valid discounts to
// a price row as json column.
func (router PaywallRouter) RefreshPriceOffers(w http.ResponseWriter, req *http.Request) {
	priceID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Find the price for this discount first.
	ftcPrice, err := router.PaywallRepo.RetrieveFtcPrice(priceID, router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// Update offers
	ftcPrice, err = router.ProductRepo.RefreshFtcPriceOffers(ftcPrice)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(ftcPrice)
}

func (router PaywallRouter) ActivatePrice(w http.ResponseWriter, req *http.Request) {
	priceID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	ftcPrice, err := router.PaywallRepo.RetrieveFtcPrice(priceID, router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	ftcPrice = ftcPrice.Activate()

	err = router.ProductRepo.ActivatePrice(ftcPrice)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(ftcPrice)
}

func (router PaywallRouter) ArchivePrice(w http.ResponseWriter, req *http.Request) {
	priceID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	p, err := router.PaywallRepo.RetrieveFtcPrice(priceID, router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	p = p.Archive()

	err = router.ProductRepo.ArchivePrice(p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	err = router.ProductRepo.ArchivePriceDiscounts(p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(p)
}