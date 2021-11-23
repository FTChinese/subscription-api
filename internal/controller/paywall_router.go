package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/products"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"net/http"
)

// PaywallRouter handles pricing plans.
type PaywallRouter struct {
	repo   products.Env
	logger *zap.Logger
	live   bool // TODO: initialize this field.

}

// NewPaywallRouter creates a new instance of pricing router.
func NewPaywallRouter(dbs db.ReadWriteMyDBs, logger *zap.Logger, c *cache.Cache, live bool) PaywallRouter {
	return PaywallRouter{
		repo:   products.NewEnv(dbs, c),
		logger: logger,
		live:   live,
	}
}

func getParamLiveMode(req *http.Request) bool {
	liveMode, err := gorest.GetQueryParam(req, "live").ToBool()
	// For backward compatibility. Query parameter
	// `live` does not exist prior to v2.6.x
	if err != nil {
		return true
	}

	return liveMode
}

// LoadPaywall loads paywall data from db or cache.
func (router PaywallRouter) LoadPaywall(w http.ResponseWriter, req *http.Request) {
	liveMode := getParamLiveMode(req)

	pw, err := router.repo.LoadPaywall(liveMode)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, pw)
}

func (router PaywallRouter) BustCache(w http.ResponseWriter, req *http.Request) {
	router.repo.ClearCache()

	pwLive, err := router.repo.LoadPaywall(true)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	pwTest, err := router.repo.LoadPaywall(false)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, map[string]pw.Paywall{
		"live": pwLive,
		"test": pwTest,
	})
}

func (router PaywallRouter) LoadPricing(w http.ResponseWriter, req *http.Request) {
	live := getParamLiveMode(req)
	p, err := router.repo.ListActivePrices(live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, p)
}
