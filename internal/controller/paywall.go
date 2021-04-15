package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/products"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"net/http"
)

// PaywallRouter handles pricing plans.
type PaywallRouter struct {
	repo   products.Env
	logger *zap.Logger
}

// NewPaywallRouter creates a new instance of pricing router.
func NewPaywallRouter(dbs db.ReadWriteSplit, c *cache.Cache, logger *zap.Logger) PaywallRouter {
	return PaywallRouter{
		repo:   products.NewEnv(dbs, c),
		logger: logger,
	}
}

// LoadPaywall loads paywall data from db or cache.
func (router PaywallRouter) LoadPaywall(w http.ResponseWriter, req *http.Request) {
	pw, err := router.repo.LoadPaywall()
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, pw)
}

func (router PaywallRouter) BustCache(w http.ResponseWriter, req *http.Request) {
	router.repo.ClearCache()

	pw, err := router.repo.LoadPaywall()
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, pw)
}

func (router PaywallRouter) LoadPricing(w http.ResponseWriter, req *http.Request) {
	p, err := router.repo.ActivePricesFromCacheOrDB()
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, p)
}
