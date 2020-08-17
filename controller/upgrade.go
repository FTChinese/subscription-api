package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/view"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/pkg/product"
	"net/http"
)

type UpgradeRouter struct {
	PayRouter
}

func NewUpgradeRouter(baseRouter PayRouter) UpgradeRouter {
	return UpgradeRouter{
		PayRouter: baseRouter,
	}
}

func (router UpgradeRouter) UpgradeBalance(w http.ResponseWriter, req *http.Request) {
	userID, _ := GetUserID(req.Header)

	pi, err := router.subEnv.PreviewUpgrade(userID)

	if err != nil {
		switch err {
		case subscription.ErrUpgradeInvalid:
			_ = view.Render(w, view.NewBadRequest(err.Error()))
		default:
			_ = view.Render(w, view.NewDBFailure(err))
		}

		return
	}

	_ = view.Render(w, view.NewResponse().SetBody(pi))
}

func (router UpgradeRouter) FreeUpgrade(w http.ResponseWriter, req *http.Request) {
	userID, _ := GetUserID(req.Header)

	p, _ := router.prodRepo.PlanByEdition(product.Edition{
		Tier:  enum.TierPremium,
		Cycle: enum.CycleYear,
	})
	clientApp := util.NewClientApp(req)

	builder := subscription.NewOrderBuilder(userID).
		SetPlan(p).
		SetClient(clientApp).
		SetEnvironment(router.subEnv.Live())

	confirmed, err := router.subEnv.FreeUpgrade(builder)
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

		return
	}

	// Save snapshot.
	go func() {
		_ = router.subEnv.BackUpMember(confirmed.Snapshot)
	}()

	orderClient := builder.ClientApp()
	// Save client app info
	go func() {
		_ = router.subEnv.SaveOrderClient(orderClient)
	}()

	wallet := builder.GetWallet()
	go func() {
		_ = router.sendFreeUpgradeEmail(confirmed.Order, wallet)
	}()

	_ = view.Render(w, view.NewNoContent())
}
