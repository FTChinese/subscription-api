package api

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
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

	prices, err := router.productRepo.ListProductPrices(productID, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(prices)
}

// update stripe price based on ftc price settings.
// This should be used when
// - creating ftc price;
// - updating ftc price;
// - attaching ftc intro price to product.
func (router PaywallRouter) updateStripPriceMeta(ftcPrice price.FtcPrice) (price.StripePrice, error) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	sugar.Infof("Updating stripe price meta")

	rawPrice, err := router.
		stripeRepo.
		Client.
		SetPriceMeta(
			ftcPrice.StripePriceID,
			price.BuildStripePriceMeta(ftcPrice))

	if err != nil {
		return price.StripePrice{}, err
	}

	p := price.NewStripePrice(rawPrice)

	err = router.stripeRepo.UpsertPrice(p)
	if err != nil {
		sugar.Error(err)
	}

	return price.NewStripePrice(rawPrice), nil
}

func (router PaywallRouter) ensureStripePrice(id string) (price.StripePrice, error) {
	sp, err := router.stripeRepo.LoadOrFetchPrice(id, false)
	if err != nil {
		return price.StripePrice{}, err
	}

	if sp.IsFromStripe {
		err := router.stripeRepo.UpsertPrice(sp)
		if err != nil {
			return price.StripePrice{}, err
		}
	}

	return sp, nil
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
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var params price.FtcCreationParams
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

	p := price.New(params, router.live)

	err := router.productRepo.CreatePrice(p)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Sync stripe price metadata
	if p.StripePriceID != "" {
		go func() {
			_, _ = router.updateStripPriceMeta(p)
		}()
	}

	// Sync legacy table.
	if p.IsRecurring() {
		go func() {
			err := router.productRepo.CreatePlan(price.NewPlan(p))
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
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	id, _ := xhttp.GetURLParam(req, "id").ToString()

	var params price.FtcUpdateParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	ftcPrice, err := router.paywallRepo.RetrievePaywallPrice(id, router.live)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	updated := ftcPrice.FtcPrice.Update(params)
	err = router.productRepo.UpdatePrice(updated)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if params.StripePriceID != "" {
		go func() {
			_, _ = router.updateStripPriceMeta(updated)
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
	ftcPrice, err := router.paywallRepo.RetrievePaywallPrice(priceID, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if ftcPrice.Kind == price.KindOneTime {
		_ = render.New(w).NotFound("Only recurring price has discounts")
		return
	}

	// Update offers
	ftcPrice, err = router.productRepo.RefreshPriceOffers(ftcPrice)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(ftcPrice)
}

// ActivatePrice flags a price to active state.
// Returns price.Price.
func (router PaywallRouter) ActivatePrice(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	priceID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		sugar.Error(err)
		return
	}

	// Load this price.
	pwPrice, err := router.paywallRepo.RetrievePaywallPrice(priceID, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		sugar.Error(err)
		return
	}

	// Check if stripe price present.
	_, err = router.ensureStripePrice(pwPrice.StripePriceID)
	if err != nil {
		sugar.Error(err)
	}
	if err != nil {
		_ = xhttp.HandleSubsErr(w, err)
		sugar.Error(err)
		return
	}

	activated := pwPrice.FtcPrice.Activate()

	err = router.productRepo.ActivatePrice(activated)
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

	p, err := router.paywallRepo.RetrievePaywallPrice(priceID, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	archived := p.FtcPrice.Archive()

	err = router.productRepo.ArchivePrice(archived)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// If this price has not discount, stop.
	if len(p.Offers) == 0 {
		_ = render.New(w).OK(archived)
		return
	}

	err = router.productRepo.ArchivePriceDiscounts(p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// The offers are already removed from price.
	// Simply return the price without offers.
	_ = render.New(w).OK(archived)
}
