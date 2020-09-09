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
	readerIDs := getReaderIDs(req.Header)

	plan, err := router.prodRepo.PlanByEdition(product.NewPremiumEdition())
	if err != nil {
		_ = render.New(w).DBError(err)
	}

	pi, err := router.subRepo.PreviewUpgrade(readerIDs, plan)

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
	defer logger.Sync()
	sugar := logger.Sugar()

	readerIDs := getReaderIDs(req.Header)
	clientApp := client.NewClientApp(req)

	p, _ := router.prodRepo.PlanByEdition(product.NewPremiumEdition())

	isTest := router.isTestAccount(readerIDs, req)

	builder := subs.NewOrderBuilder(readerIDs).
		SetPlan(p).
		SetTest(isTest)

	confirmed, err := router.subRepo.FreeUpgrade(builder)
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
		_ = router.readerRepo.BackUpMember(confirmed.Snapshot)
	}()

	// Save client app info
	go func() {
		_ = router.subRepo.SaveOrderClient(client.OrderClient{
			OrderID: confirmed.Order.ID,
			Client:  clientApp,
		})
	}()

	// Send email
	go func() {
		// Find this user's personal data
		account, err := router.readerRepo.AccountByFtcID(confirmed.Order.FtcID.String)
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
			sugar.Error(err)
			return
		}
	}()

	_ = render.New(w).NoContent()
}
