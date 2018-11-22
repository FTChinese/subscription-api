package controller

import (
	"net/http"

	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/util"
)

// DefaultPlans shows what our subscription plans are.
func DefaultPlans() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		util.Render(w, util.NewResponse().NoCache().SetBody(model.DefaultPlans))
	}
}

// PaywallRouter handles pricing plans.
type PaywallRouter struct {
	model model.Env
}

// NewPaywallRouter creates a new instance of pricing router.
func NewPaywallRouter(m model.Env) PaywallRouter {
	return PaywallRouter{model: m}
}

// RefreshSchedule busts cache and retrieve a latest schedule if exists.
func (pr PaywallRouter) RefreshSchedule(w http.ResponseWriter, req *http.Request) {
	sch, err := pr.model.RetrieveSchedule()

	if err != nil {
		util.Render(w, util.NewDBFailure(err))

		return
	}

	util.Render(w, util.NewResponse().NoCache().SetBody(sch))
}

// DiscountSchedule shows what schedule is being used currently.
func (pr PaywallRouter) DiscountSchedule(w http.ResponseWriter, req *http.Request) {
	sch, found := pr.model.ScheduleFromCache()

	if !found {
		util.Render(w, util.NewNotFound())

		return
	}

	util.Render(w, util.NewResponse().NoCache().SetBody(sch))
}

// CurrentPlans show what plans are being used after merging with discount plans.
func (pr PaywallRouter) CurrentPlans(w http.ResponseWriter, req *http.Request) {
	plans := pr.model.GetCurrentPlans()

	util.Render(w, util.NewResponse().NoCache().SetBody(plans))
}
