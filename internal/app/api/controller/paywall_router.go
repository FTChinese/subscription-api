package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/internal/repository/products"
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"go.uber.org/zap"
	"net/http"
)

// PaywallRouter handles pricing plans.
type PaywallRouter struct {
	ProductRepo     products.Env
	PaywallRepo     shared.PaywallCommon
	StripePriceRepo shared.StripeBaseRepo
	Logger          *zap.Logger
	Live            bool
}

// NewPaywallRouter creates a new instance of pricing router.
func NewPaywallRouter(
	prodRepo products.Env,
	paywallBaseRepo shared.PaywallCommon,
	stripeBaseRepo shared.StripeBaseRepo,
	logger *zap.Logger,
	live bool,
) PaywallRouter {
	return PaywallRouter{
		ProductRepo:     prodRepo,
		StripePriceRepo: stripeBaseRepo,
		PaywallRepo:     paywallBaseRepo,
		Logger:          logger,
		Live:            live,
	}
}

// LoadPaywall loads paywall data from db or cache.
func (router PaywallRouter) LoadPaywall(w http.ResponseWriter, req *http.Request) {

	paywall, err := router.PaywallRepo.LoadPaywall(router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, paywall)
}

// BustCache clears the cached paywall data.
func (router PaywallRouter) BustCache(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	router.PaywallRepo.ClearCache()

	paywall, err := router.PaywallRepo.LoadPaywall(router.Live)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	_, err = router.StripePriceRepo.ListPrices(router.Live, true)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	stripeIDs := paywall.StripePriceIDs()

	for _, id := range stripeIDs {
		_, err := router.StripePriceRepo.LoadPrice(id, false)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).DBError(err)
			return
		}
	}

	// List stripe prices from what we have just cached.
	stripePrices, err := router.StripePriceRepo.ListPrices(router.Live, false)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, struct {
		Paywall      pw.Paywall     `json:"paywall"`
		StripePrices []stripe.Price `json:"stripePrices"`
	}{
		Paywall:      paywall,
		StripePrices: stripePrices,
	})
}

func (router PaywallRouter) LoadPricing(w http.ResponseWriter, req *http.Request) {
	p, err := router.PaywallRepo.ListActivePrices(router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, p)
}