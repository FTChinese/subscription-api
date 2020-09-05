package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/client"
	"github.com/FTChinese/subscription-api/pkg/letter"
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
	userID := getReaderIDs(req.Header)

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

// FreeUpgrade handles free upgrade request.
func (router UpgradeRouter) FreeUpgrade(w http.ResponseWriter, req *http.Request) {
	userID := getReaderIDs(req.Header)

	p, _ := router.prodRepo.PlanByEdition(product.NewPremiumEdition())
	clientApp := client.NewClientApp(req)

	builder := subs.NewOrderBuilder(userID).
		SetPlan(p).
		SetEnvConfig(router.config)

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
		_ = router.readerEnv.BackUpMember(confirmed.Snapshot)
	}()

	// Save client app info
	go func() {
		_ = router.subEnv.SaveOrderClient(client.OrderClient{
			OrderID: confirmed.Order.ID,
			Client:  clientApp,
		})
	}()

	// Send email
	go func() {
		// Find this user's personal data
		account, err := router.readerEnv.AccountByFtcID(confirmed.Order.FtcID.String)
		if err != nil {
			return
		}
		parcel, err := letter.NewFreeUpgradeParcel(
			account,
			confirmed.Order,
			builder.GetWallet().Sources,
		)

		err = router.postman.Deliver(parcel)
		if err != nil {
			logger.Error(err)
			return
		}
	}()

	_ = render.New(w).NoContent()
}
