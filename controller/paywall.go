package controller

import (
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"net/http"

	"github.com/FTChinese/go-rest/view"
	"gitlab.com/ftchinese/subscription-api/repository"
)

// PaywallRouter handles pricing plans.
type PaywallRouter struct {
	model repository.Env
}

// NewPaywallRouter creates a new instance of pricing router.
func NewPaywallRouter(m repository.Env) PaywallRouter {
	return PaywallRouter{
		model: m,
	}
}

// GetPaywall loads current paywall in effect.
func (router PaywallRouter) GetPaywall(w http.ResponseWriter, req *http.Request) {
	pw, err := router.model.GetPayWall()
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
		paywall.GetFtcPlans(true))

	if err != nil {
		_ = view.Render(w, view.NewInternalError(err.Error()))
		return
	}

	_ = view.Render(w, view.NewResponse().SetBody(pw))
}

// GetPricing loads current pricing plans in effect.
func (router PaywallRouter) GetPricing(w http.ResponseWriter, req *http.Request) {
	p := router.model.GetCurrentPlans()

	view.Render(w, view.NewResponse().SetBody(p))
}

// DefaultPlans shows what our subscription plans are.
func DefaultPricing(w http.ResponseWriter, req *http.Request) {

	_ = view.Render(
		w,
		view.NewResponse().
			NoCache().
			SetBody(paywall.GetFtcPlans(true)))
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

// RefreshPromo busts cache and retrieve a latest promotion schedule if exists.
// The retrieved promotion is put into cache and also send back to the request.
func (router PaywallRouter) RefreshPromo(w http.ResponseWriter, req *http.Request) {
	promo, err := router.model.RetrievePromo()

	if err != nil {
		view.Render(w, view.NewDBFailure(err))

		return
	}

	view.Render(w, view.NewResponse().SetBody(promo))
}
