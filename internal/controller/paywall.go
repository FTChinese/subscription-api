package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/products"
	"github.com/FTChinese/subscription-api/internal/repository/subrepo"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"net/http"
)

// PaywallRouter handles pricing plans.
type PaywallRouter struct {
	env  subrepo.Env
	repo products.Env
}

// NewPaywallRouter creates a new instance of pricing router.
func NewPaywallRouter(db *sqlx.DB, c *cache.Cache, b config.BuildConfig) PaywallRouter {
	return PaywallRouter{
		env:  subrepo.NewEnv(db, c, b),
		repo: products.NewEnv(db, c),
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
	p, err := router.repo.LoadPricing()
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, p)
}
