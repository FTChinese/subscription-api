package controller

import (
	"database/sql"
	"net/http"

	"github.com/FTChinese/go-rest/view"
	cache "github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

// PaywallRouter handles pricing plans.
type PaywallRouter struct {
	model model.Env
}

// NewPaywallRouter creates a new instance of pricing router.
func NewPaywallRouter(db *sql.DB, c *cache.Cache) PaywallRouter {
	return PaywallRouter{
		model: model.Env{
			DB:    db,
			Cache: c,
		},
	}
}

// GetPromo gets the current effective promotion schedule.
func (router PaywallRouter) GetPromo(w http.ResponseWriter, req *http.Request) {
	promo, found := router.model.LoadCachedPromo()

	if !found {
		view.Render(w, view.NewNotFound())

		return
	}

	view.Render(w, view.NewResponse().NoCache().SetBody(promo))
}

// DefaultPlans shows what our subscription plans are.
func DefaultPlans(w http.ResponseWriter, req *http.Request) {

	view.Render(
		w,
		view.NewResponse().
			NoCache().
			SetBody(paywall.GetDefaultPricing()))
}
