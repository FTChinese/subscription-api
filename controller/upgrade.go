package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/client"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/subs"
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

	plan, err := router.prodRepo.PlanByEdition(product.NewPremiumEdition())
	if err != nil {
		_ = render.New(w).DBError(err)
	}

	pi, err := router.subEnv.PreviewUpgrade(userID, plan)

	if err != nil {
		switch err {
		case subs.ErrUpgradeInvalid:
			_ = render.New(w).BadRequest(err.Error())
		default:
			_ = render.New(w).DBError(err)
		}
		return
	}

	_ = render.New(w).JSON(http.StatusOK, pi)
}

func (router UpgradeRouter) FreeUpgrade(w http.ResponseWriter, req *http.Request) {
	userID, _ := GetUserID(req.Header)

	p, _ := router.prodRepo.PlanByEdition(product.NewPremiumEdition())
	clientApp := client.NewClientApp(req)

	builder := subs.NewOrderBuilder(userID).
		SetPlan(p).
		SetEnvironment(router.subEnv.Live())

	confirmed, err := router.subEnv.FreeUpgrade(builder)
	if err != nil {
		switch err {
		case subs.ErrUpgradeInvalid:
			_ = render.New(w).BadRequest(err.Error())

		case subs.ErrBalanceCannotCoverUpgrade:
			pi, _ := builder.PaymentIntent()
			_ = render.New(w).JSON(http.StatusOK, pi)

		default:
			_ = render.New(w).DBError(err)
		}
		return
	}

	// Save snapshot.
	go func() {
		_ = router.subEnv.BackUpMember(confirmed.Snapshot)
	}()

	// Save client app info
	go func() {
		_ = router.subEnv.SaveOrderClient(client.OrderClient{
			OrderID: confirmed.Order.ID,
			Client:  clientApp,
		})
	}()

	wallet := builder.GetWallet()
	go func() {
		_ = router.sendFreeUpgradeEmail(confirmed.Order, wallet)
	}()

	_ = render.New(w).NoContent()
}
