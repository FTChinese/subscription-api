package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/products"
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"go.uber.org/zap"
	"net/http"
)

// PaywallRouter handles pricing plans.
type PaywallRouter struct {
	prodRepo   products.Env
	stripeRepo shared.StripeBaseRepo
	logger     *zap.Logger
	live       bool
}

// NewPaywallRouter creates a new instance of pricing router.
func NewPaywallRouter(
	paywallBaseRepo shared.PaywallCommon,
	stripeBaseRepo shared.StripeBaseRepo,
	logger *zap.Logger,
	live bool,
) PaywallRouter {
	return PaywallRouter{
		prodRepo:   products.New(paywallBaseRepo),
		stripeRepo: stripeBaseRepo,
		logger:     logger,
		live:       live,
	}
}

// LoadPaywall loads paywall data from db or cache.
func (router PaywallRouter) LoadPaywall(w http.ResponseWriter, req *http.Request) {

	paywall, err := router.prodRepo.LoadPaywall(router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, paywall)
}

// BustCache clears the cached paywall data.
func (router PaywallRouter) BustCache(w http.ResponseWriter, req *http.Request) {
	router.prodRepo.ClearCache()

	paywall, err := router.prodRepo.LoadPaywall(router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_, err = router.stripeRepo.ListPrices(router.live, true)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	stripeIDs := paywall.StripePriceIDs()

	for _, id := range stripeIDs {
		_, err := router.stripeRepo.LoadPrice(id, false)
		if err != nil {
			_ = render.New(w).DBError(err)
			return
		}
	}

	stripePrices, err := router.stripeRepo.ListPrices(router.live, false)
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
	p, err := router.prodRepo.ListActivePrices(router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, p)
}
