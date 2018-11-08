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

// PricingRouter handles pricing plans.
type PricingRouter struct {
	model model.Env
}

// NewPricingRouter creates a new instance of pricing router.
func NewPricingRouter(m model.Env) PricingRouter {
	return PricingRouter{model: m}
}

// RefreshSchedule busts cache and retrieve a latest schedule if exists.
func (pr PricingRouter) RefreshSchedule(w http.ResponseWriter, req *http.Request) {
	sch, err := pr.model.RetrieveSchedule()

	if err != nil {
		util.Render(w, util.NewDBFailure(err))

		return
	}

	util.Render(w, util.NewResponse().NoCache().SetBody(sch))
}

// DiscountSchedule shows what schedule is being used currently.
func (pr PricingRouter) DiscountSchedule(w http.ResponseWriter, req *http.Request) {
	sch, found := pr.model.ScheduleFromCache()

	if !found {
		util.Render(w, util.NewNotFound())

		return
	}

	util.Render(w, util.NewResponse().NoCache().SetBody(sch))
}

// CurrentPlans show what plans are being used after merging with discount plans.
func (pr PricingRouter) CurrentPlans(w http.ResponseWriter, req *http.Request) {
	plans := pr.model.GetCurrentPlans()

	util.Render(w, util.NewResponse().NoCache().SetBody(plans))
}
