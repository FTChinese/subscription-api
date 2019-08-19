package controller

import (
	"github.com/FTChinese/go-rest/view"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository"
	"net/http"
	"time"
)

type UpgradeRouter struct {
	PayRouter
}

func NewUpgradeRouter(m repository.Env) UpgradeRouter {
	r := UpgradeRouter{}
	r.model = m

	return r
}

// PreviewUpgrade calculates the proration of active or
// unused orders.
// Response:
//
// 404 membership not found
// 204 already a premium member
// 422 field: membership, code: already_upgraded
// Deprecate
func (router UpgradeRouter) PreviewUpgrade(w http.ResponseWriter, req *http.Request) {
	user, _ := GetUserID(req.Header)

	balance, err := router.model.UpgradeBalance(user)
	if err != nil {
		switch err {
		case util.ErrAlreadyUpgraded:
			r := view.NewReason()
			r.Field = "membership"
			r.Code = "already_upgraded"
			r.SetMessage("Membership is already premium")
			view.Render(w, view.NewUnprocessable(r))
			return
		}

		// membership not found is handled here.
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// Tell client how much user should pay for upgrading.
	view.Render(w, view.NewResponse().SetBody(balance))
}

func (router UpgradeRouter) UpgradeBalance(w http.ResponseWriter, req *http.Request) {
	userID, _ := GetUserID(req.Header)

	up, err := router.model.PreviewUpgrade(userID)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	view.Render(w, view.NewResponse().SetBody(up))
}

// DirectUpgrade performs membership upgrading for users whose
// account balance could cover the upgrading expense exactly.
// 404 if membership is not found
// 422 if already a premium
// 200 if balance is not enough to cover upgrading cost.
// 204 if upgraded successfully.
// Deprecate
func (router UpgradeRouter) DirectUpgrade(w http.ResponseWriter, req *http.Request) {
	user, _ := GetUserID(req.Header)

	upgradePlan, err := router.model.UpgradeBalance(user)
	if err != nil {
		switch err {
		case util.ErrAlreadyUpgraded:
			r := view.NewReason()
			r.Field = "membership"
			r.Code = "already_upgraded"
			r.SetMessage("Membership is already premium")
			view.Render(w, view.NewUnprocessable(r))
			return
		}

		// membership not found is handled here.
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// If user needs to pay any extra money.
	if upgradePlan.Payable > 0 {
		view.Render(w, view.NewResponse().SetBody(upgradePlan))
		return
	}

	subs, err := router.model.DirectUpgradeOrder(user, upgradePlan, util.NewClientApp(req))
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// Confirm this order
	updatedSubs, err := router.model.ConfirmPayment(subs.ID, time.Now())
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	go router.sendConfirmationEmail(updatedSubs)

	view.Render(w, view.NewNoContent())
}

func (router UpgradeRouter) FreeUpgrade(w http.ResponseWriter, req *http.Request) {
	userID, _ := GetUserID(req.Header)

	up, err := router.model.PreviewUpgrade(userID)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// If user needs to pay any extra money.
	if up.Plan.NetPrice > 0 {
		view.Render(w, view.NewResponse().SetBody(up))
		return
	}

	subs, err := router.model.FreeUpgrade(userID, up, util.NewClientApp(req))
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	go func() {
		err := router.sendConfirmationEmail(subs)
		if err != nil {
			logrus.WithField("trace", "UpgradeRouter.FreeUpgrade").Error(err)
		}
	}()

	view.Render(w, view.NewNoContent())
}
