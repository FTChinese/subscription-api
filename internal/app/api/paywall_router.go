package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/products"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"net/http"
)

// PaywallRouter handles pricing plans.
type PaywallRouter struct {
	productRepo products.Env
	PaymentShared
}

func NewPaywallRouter(
	dbs db.ReadWriteMyDBs,
	c *cache.Cache,
	logger *zap.Logger,
	live bool,
) PaywallRouter {
	return PaywallRouter{
		productRepo: products.Env{},
		PaymentShared: NewPaymentShared(
			dbs,
			c,
			logger,
			live),
	}
}

// LoadPaywall loads paywall data from db or cache.
func (router PaywallRouter) LoadPaywall(w http.ResponseWriter, req *http.Request) {

	refresh := xhttp.ParseQueryRefresh(req)

	paywall, err := router.LoadCachedPaywall(refresh)

	if err != nil {
		_ = xhttp.HandleStripeErr(w, err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, paywall)
}

// BustCache clears the cached paywall data.
// Deprecated
func (router PaywallRouter) BustCache(w http.ResponseWriter, _ *http.Request) {

	paywall, err := router.LoadCachedPaywall(true)
	if err != nil {
		_ = xhttp.HandleStripeErr(w, err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, paywall)
}

func (router PaywallRouter) LoadPricing(w http.ResponseWriter, _ *http.Request) {
	p, err := router.paywallRepo.ListActivePrices(router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, p)
}
