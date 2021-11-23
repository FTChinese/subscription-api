package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
	"net/http"
)

// ListPrices retrieves all prices under a product.
func (router PaywallRouter) ListPrices(w http.ResponseWriter, req *http.Request) {
	productID, err := gorest.GetQueryParam(req, "product_id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	prices, err := router.repo.ListPrices(productID, router.live)
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

	p := price.NewFtcPrice(params, router.live)

	err := router.repo.CreatePrice(p)
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
	id, _ := getURLParam(req, "id").ToString()

	var params price.FtcPriceUpdateParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	ftcPrice, err := router.repo.RetrieveFtcPrice(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	updated := ftcPrice.WithUpdate(params)
	err = router.repo.UpdateFtcPrice(updated)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(updated)
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
