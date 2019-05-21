package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/view"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
	"net/http"
	"time"
)

type UpgradeRouter struct {
	PayRouter
}

func NewUpgradeRouter(m model.Env) UpgradeRouter {
	r := UpgradeRouter{}
	r.model = m

	return r
}

// PreviewUpgrade calculates the proration of active or
// unused orders.
func (router UpgradeRouter) PreviewUpgrade(w http.ResponseWriter, req *http.Request) {
	user, _ := GetUser(req.Header)

	// Retrieve this user's current membership.
	// If not found, deny upgrading.
	member, err := router.model.RetrieveMember(user)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// If user is already a premium member, do nothing.
	if member.Tier == enum.TierPremium {
		view.Render(w, view.NewNoContent())
		return
	}

	// Retrieve unused portion of orders
	orders, err := router.model.FindProration(user)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// Find the current plan for yearly premium.
	plan, _ := router.model.GetCurrentPricing().
		FindPlan(
			enum.TierPremium.String(),
			enum.CycleYear.String())

	up := paywall.NewUpgradePlan(plan).
		SetProration(orders).
		CalculatePayable()

	// Tell client how much user should pay for upgrading.
	view.Render(w, view.NewResponse().SetBody(up))
}

// DirectUpgrade performs membership upgrading for users whose
// account balance could cover the upgrading expense exactly.
func (router UpgradeRouter) DirectUpgrade(w http.ResponseWriter, req *http.Request) {
	user, _ := GetUser(req.Header)

	// Retrieve this user's current membership.
	// If not found, deny upgrading.
	member, err := router.model.RetrieveMember(user)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// If user is already a premium member, do nothing.
	if member.Tier == enum.TierPremium {
		view.Render(w, view.NewNoContent())
		return
	}

	// Find the current plan for yearly premium.
	plan, err := router.model.GetCurrentPricing().
		FindPlan(
			enum.TierPremium.String(),
			enum.CycleYear.String())

	if err != nil {
		logger.WithField("trace", "DirectUpgrade").Error(err)
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	up, err := router.model.BuildUpgradePlan(user, plan)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// If user needs to pay any extra money.
	if up.Payable > 0 {
		view.Render(w, view.NewResponse().SetBody(up))
		return
	}

	// If user do not need to pay, upgrade directly.
	// Create an order whose net price is 0.
	subs, err := paywall.NewSubsUpgrade(user, up)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	err = router.model.SaveSubscription(subs, util.NewClientApp(req))
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// Update this member.
	updatedSubs, err := router.model.ConfirmPayment(subs.OrderID, time.Now())
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	go router.sendConfirmationEmail(updatedSubs)

	view.Render(w, view.NewNoContent())
}
