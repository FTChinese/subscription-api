package controller

import (
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/pkg/config"
	"gitlab.com/ftchinese/subscription-api/repository/subrepo"
	"net/http"

	"github.com/FTChinese/go-rest/view"
)

// PaywallRouter handles pricing plans.
type PaywallRouter struct {
	env subrepo.SubEnv
}

// NewPaywallRouter creates a new instance of pricing router.
func NewPaywallRouter(db *sqlx.DB, c *cache.Cache, b config.BuildConfig) PaywallRouter {
	return PaywallRouter{
		env: subrepo.NewSubEnv(db, c, b),
	}
}

// GetPaywall loads current paywall in effect.
func (router PaywallRouter) GetPaywall(w http.ResponseWriter, req *http.Request) {
	pw, err := router.env.GetPayWall()
	if err != nil {
		view.Render(w, view.NewInternalError(err.Error()))
		return
	}

	view.Render(w, view.NewResponse().SetBody(pw))
}

// DefaultPaywall loads default paywall data.
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
func (router PaywallRouter) GetPricing(w http.ResponseWriter, req *http.Request) {
	p := router.env.GetCurrentPlans()

	view.Render(w, view.NewResponse().SetBody(p))
}

// DefaultPlans shows what our subscription plans are.
func DefaultPricing(w http.ResponseWriter, req *http.Request) {

	_ = view.Render(
		w,
		view.NewResponse().
			NoCache().
			SetBody(plan.GetPlans()))
}

// GetPromo gets the current effective promotion schedule.
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
func (router PaywallRouter) RefreshPromo(w http.ResponseWriter, req *http.Request) {
	promo, err := router.env.RetrievePromo()

	if err != nil {
		view.Render(w, view.NewDBFailure(err))

		return
	}

	view.Render(w, view.NewResponse().SetBody(promo))
}
