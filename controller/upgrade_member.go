package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/view"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository/subrepo"
	"net/http"
)

type UpgradeRouter struct {
	PayRouter
}

func NewUpgradeRouter(env subrepo.SubEnv) UpgradeRouter {
	r := UpgradeRouter{}
	r.subEnv = env

	return r
}

func (router UpgradeRouter) UpgradeBalance(w http.ResponseWriter, req *http.Request) {
	userID, _ := GetUserID(req.Header)

	p, _ := plan.FindPlan(enum.TierPremium, enum.CycleYear)

	builder := subscription.NewOrderBuilder(userID).
		SetPlan(p).SetEnvironment(router.subEnv.Live())

	otx, err := router.subEnv.BeginOrderTx()
	if err != nil {
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	err = otx.PreviewUpgrade(builder)
	if err != nil {
		switch err {
		case subscription.ErrUpgradeInvalid:
			_ = view.Render(w, view.NewBadRequest(err.Error()))
		default:
			_ = view.Render(w, view.NewDBFailure(err))
		}

		_ = otx.Rollback()
		return
	}

	if err := otx.Commit(); err != nil {
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	pi, err := builder.PaymentIntent()

	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	_ = view.Render(w, view.NewResponse().SetBody(pi))
}

func (router UpgradeRouter) FreeUpgrade(w http.ResponseWriter, req *http.Request) {
	userID, _ := GetUserID(req.Header)

	p, _ := plan.FindPlan(enum.TierPremium, enum.CycleYear)
	clientApp := util.NewClientApp(req)

	builder := subscription.NewOrderBuilder(userID).
		SetPlan(p).
		SetClient(clientApp).
		SetEnvironment(router.subEnv.Live())

	otx, err := router.subEnv.BeginOrderTx()
	if err != nil {
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	order, err := otx.FreeUpgrade(builder)
	if err != nil {
		switch err {
		case subscription.ErrUpgradeInvalid:
			_ = view.Render(w, view.NewBadRequest(err.Error()))

		case subscription.ErrBalanceCannotCoverUpgrade:
			pi, _ := builder.PaymentIntent()
			_ = view.Render(w, view.NewResponse().SetBody(pi))

		default:
			_ = view.Render(w, view.NewDBFailure(err))
		}

		_ = otx.Rollback()
		return
	}

	if err := otx.Commit(); err != nil {
		return
	}

	orderClient := builder.ClientApp()
	// Save client app info
	go func() {
		_ = router.subEnv.SaveOrderClient(orderClient)
	}()

	snapshot := builder.MembershipSnapshot()
	go func() {
		_ = router.subEnv.BackUpMember(snapshot)
	}()

	go func() {
		_ = router.sendConfirmationEmail(order)
	}()

	_ = view.Render(w, view.NewNoContent())
}
