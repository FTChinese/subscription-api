package controller

import (
	"net/http"

	"github.com/FTChinese/go-rest/view"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

// PaywallRouter handles pricing plans.
type PaywallRouter struct {
	model model.Env
}

// NewPaywallRouter creates a new instance of pricing router.
func NewPaywallRouter(m model.Env) PaywallRouter {
	return PaywallRouter{
		model: m,
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
