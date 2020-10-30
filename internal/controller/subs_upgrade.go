package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/client"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"net/http"
)

func (router SubsRouter) PreviewUpgrade(w http.ResponseWriter, req *http.Request) {
	readerIDs := getReaderIDs(req.Header)

	account, err := router.readerRepo.FindAccount(readerIDs)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	plan, err := router.prodRepo.PlanByEdition(product.NewPremiumEdition())
	if err != nil {
		_ = render.New(w).DBError(err)
	}

	pi, err := router.subRepo.UpgradeIntent(account, plan, true)

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
func (router SubsRouter) FreeUpgrade(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	readerIDs := getReaderIDs(req.Header)
	account, err := router.readerRepo.FindAccount(readerIDs)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	clientApp := client.NewClientApp(req)

	p, _ := router.prodRepo.PlanByEdition(product.NewPremiumEdition())

	intent, err := router.subRepo.UpgradeIntent(account, p, false)
	// Check whether intent if free. If not free, return it to client and stop.
	if err != nil {
		switch err {
		case subs.ErrUpgradeInvalid:
			_ = render.New(w).BadRequest(err.Error())

		default:
			_ = render.New(w).DBError(err)
		}
		return
	}

	if !intent.IsFree {
		_ = render.New(w).OK(intent)
		return
	}

	// Only free upgrade could go to here.
	// Save snapshot.
	go func() {
		_ = router.readerRepo.BackUpMember(intent.Result.Snapshot)
	}()

	// Save client app info
	go func() {
		_ = router.subRepo.LogOrderMeta(subs.OrderMeta{
			OrderID: intent.Result.Order.ID,
			Client:  clientApp,
		})
	}()

	// Send email
	go func() {
		// Find this user's personal data
		parcel, err := letter.NewFreeUpgradeParcel(
			account,
			intent.Result.Order,
			intent.Wallet.Sources,
		)

		err = router.postman.Deliver(parcel)
		if err != nil {
			sugar.Error(err)
			return
		}
	}()

	_ = render.New(w).NoContent()
}
