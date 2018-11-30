package controller

import (
	"net/http"

	"gitlab.com/ftchinese/subscription-api/util"
)

// CurrentPlans show what plans are being used after merging with discount plans.
// This is used to inspect server internal status
// so that you can see what pricing plans are being used when user begin to pay.
func (pr PaywallRouter) CurrentPlans(w http.ResponseWriter, req *http.Request) {
	plans := pr.model.GetCurrentPlans()

	util.Render(w, util.NewResponse().NoCache().SetBody(plans))
}

// RefreshPromo busts cache and retrieve a latest promotion schedule if exists.
// The retrieved promotion is put into cache and also send back to the request.
func (pr PaywallRouter) RefreshPromo(w http.ResponseWriter, req *http.Request) {
	sch, err := pr.model.RetrievePromo()

	if err != nil {
		util.Render(w, util.NewDBFailure(err))

		return
	}

	util.Render(w, util.NewResponse().NoCache().SetBody(sch))
}
