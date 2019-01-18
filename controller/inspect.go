package controller

import (
	"net/http"

	"github.com/FTChinese/go-rest/view"
)

// CurrentPlans show what plans are being used after merging with discount plans.
// This is used to inspect server internal status
// so that you can see what pricing plans are being used when user begin to pay.
func (router PaywallRouter) CurrentPlans(w http.ResponseWriter, req *http.Request) {
	plans := router.model.GetCurrentPricing()

	view.Render(w, view.NewResponse().NoCache().SetBody(plans))
}

// RefreshPromo busts cache and retrieve a latest promotion schedule if exists.
// The retrieved promotion is put into cache and also send back to the request.
func (router PaywallRouter) RefreshPromo(w http.ResponseWriter, req *http.Request) {
	promo, err := router.model.RetrievePromo()

	if err != nil {
		view.Render(w, view.NewDBFailure(err))

		return
	}

	view.Render(w, view.NewResponse().NoCache().SetBody(promo))
}
