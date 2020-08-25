package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/models/paywall"
	"github.com/FTChinese/subscription-api/models/plan"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/repository/products"
	"github.com/FTChinese/subscription-api/repository/subrepo"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"net/http"

	"github.com/FTChinese/go-rest/view"
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

}

// GetPaywall loads current paywall in effect.
// Deprecated
func (router PaywallRouter) GetPaywall(w http.ResponseWriter, req *http.Request) {
	pw, err := router.env.GetPayWall()
	if err != nil {
		view.Render(w, view.NewInternalError(err.Error()))
		return
	}

	view.Render(w, view.NewResponse().SetBody(pw))
}

// DefaultPaywall loads default paywall data.
// Deprecated
func DefaultPaywall(w http.ResponseWriter, req *http.Request) {
	pw, err := paywall.BuildPayWall(
		paywall.GetDefaultBanner(),
		plan.GetPlans())

	if err != nil {
		_ = view.Render(w, view.NewInternalError(err.Error()))
		return
	}

	_ = view.Render(w, view.NewResponse().SetBody(pw))
}

// GetPricing loads current pricing plans in effect.
// Deprecated
func (router PaywallRouter) GetPricing(w http.ResponseWriter, req *http.Request) {
	p := router.env.GetCurrentPlans()

	view.Render(w, view.NewResponse().SetBody(p))
}

// DefaultPlans shows what our subscription plans are.
// Deprecated
func DefaultPricing(w http.ResponseWriter, req *http.Request) {

	_ = view.Render(
		w,
		view.NewResponse().
			NoCache().
			SetBody(plan.GetPlans()))
}

// GetPromo gets the current effective promotion schedule.
// Deprecated
func (router PaywallRouter) GetPromo(w http.ResponseWriter, req *http.Request) {
	promo, found := router.env.LoadCachedPromo()

	if !found {
		view.Render(w, view.NewNotFound())

		return
	}

	view.Render(w, view.NewResponse().NoCache().SetBody(promo))
}

// RefreshPromo busts cache and retrieve a latest promotion schedule if exists.
// The retrieved promotion is put into cache and also send back to the request.
// Deprecated
func (router PaywallRouter) RefreshPromo(w http.ResponseWriter, req *http.Request) {
	promo, err := router.env.RetrievePromo()

	if err != nil {
		view.Render(w, view.NewDBFailure(err))

		return
	}

	view.Render(w, view.NewResponse().SetBody(promo))
}
