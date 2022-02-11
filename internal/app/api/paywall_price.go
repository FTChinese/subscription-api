package api

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// ListPrices retrieves all prices under a product.
// The price's discounts column will be included.
// Returns []pw.PaywallPrice.
func (router PaywallRouter) ListPrices(w http.ResponseWriter, req *http.Request) {
	productID, err := gorest.GetQueryParam(req, "product_id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	prices, err := router.ProductRepo.ListProductPrices(productID, router.Live)
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
// Returns price.Price.
func (router PaywallRouter) CreatePrice(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	var params price.CreationParams
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

	p := price.New(params, router.Live)

	err := router.ProductRepo.CreatePrice(p)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Sync stripe price metadata
	if p.StripePriceID != "" {
		go func() {
			sugar.Infof("Updating stripe price %s metadata", p.StripePriceID)
			sp, err := router.StripePrice.
				UpdatePriceMeta(
					params.StripePriceID,
					stripe.PriceMetaParams(p, p.IsOneTime()))

			if err != nil {
				sugar.Error(err)
				return
			}

			sugar.Infof("Stripe price meta set %v", sp)
		}()
	}

	// Sync legacy table.
	if p.IsRecurring() {
		go func() {
			err := router.ProductRepo.CreatePlan(price.NewPlan(p))
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(p)
}

// UpdatePrice changes a price's Title, Nickname, or StripePriceID fields
// Input body:
// - description?: string;
// - nickname?: string;
// - stripePriceId: string;
// Return price.Price.
func (router PaywallRouter) UpdatePrice(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	id, _ := xhttp.GetURLParam(req, "id").ToString()

	var params price.UpdateParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	ftcPrice, err := router.PaywallRepo.RetrievePaywallPrice(id, router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	updated := ftcPrice.Price.Update(params)
	err = router.ProductRepo.UpdatePrice(updated)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if params.StripePriceID != "" {
		go func() {
			sugar.Infof("Updating stripe price %s metadata", params.StripePriceID)
			sp, err := router.StripePrice.
				UpdatePriceMeta(
					params.StripePriceID,
					stripe.PriceMetaParams(
						updated,
						updated.IsOneTime() && updated.Active))

			if err != nil {
				sugar.Error(err)
				return
			}

			sugar.Infof("Stripe price meta updated %v", sp)
		}()
	}

	_ = render.New(w).OK(updated)
}

// RefreshPriceOffers attaches all valid discounts to
// a price row as json column.
// Returns pw.PaywallPrice
func (router PaywallRouter) RefreshPriceOffers(w http.ResponseWriter, req *http.Request) {
	priceID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Find the price for this discount first.
	ftcPrice, err := router.PaywallRepo.RetrievePaywallPrice(priceID, router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if ftcPrice.Kind == price.KindOneTime {
		_ = render.New(w).NotFound("Only recurring price has discounts")
		return
	}

	// Update offers
	ftcPrice, err = router.ProductRepo.RefreshPriceOffers(ftcPrice)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(ftcPrice)
}

// ActivatePrice flags a price to active state.
// Returns price.Price.
func (router PaywallRouter) ActivatePrice(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	priceID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		sugar.Error(err)
		return
	}

	// Load this price.
	pwPrice, err := router.PaywallRepo.RetrievePaywallPrice(priceID, router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		sugar.Error(err)
		return
	}

	// Check if stripe price present.
	_, err = router.StripePrice.LoadPrice(pwPrice.StripePriceID, router.Live)
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		sugar.Error(err)
		return
	}

	activated := pwPrice.Price.Activate()

	err = router.ProductRepo.ActivatePrice(activated)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// If the price is a one_time price,
	_ = render.New(w).OK(activated)
}

// ArchivePrice flags a price as deleted.
// It should never be touched after this operation.
// The returned price has no offers attached since they are
// all removed.
// Returns price.Price.
func (router PaywallRouter) ArchivePrice(w http.ResponseWriter, req *http.Request) {
	priceID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	p, err := router.PaywallRepo.RetrievePaywallPrice(priceID, router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	archived := p.Price.Archive()

	err = router.ProductRepo.ArchivePrice(archived)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// If this price has not discount, stop.
	if len(p.Offers) == 0 {
		_ = render.New(w).OK(archived)
		return
	}

	err = router.ProductRepo.ArchivePriceDiscounts(p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// The offers are already removed from price.
	// Simply return the price without offers.
	_ = render.New(w).OK(archived)
}
