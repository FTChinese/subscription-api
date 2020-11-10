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

	account, err := router.ReaderRepo.FindAccount(readerIDs)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	plan, err := router.prodRepo.PlanByEdition(product.NewPremiumEdition())
	if err != nil {
		_ = render.New(w).DBError(err)
	}

	config := subs.NewPayment(account, plan).
		WithPreview(true)

	pi, err := router.SubsRepo.UpgradeIntent(config)

	if err != nil {
		switch err {
		case subs.ErrNotUpgradeIntent:
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
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	readerIDs := getReaderIDs(req.Header)
	account, err := router.ReaderRepo.FindAccount(readerIDs)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	clientApp := client.NewClientApp(req)

	p, _ := router.prodRepo.PlanByEdition(product.NewPremiumEdition())

	config := subs.NewPayment(account, p)
	intent, err := router.SubsRepo.UpgradeIntent(config)
	// Check whether intent if free. If not free, return it to client and stop.
	if err != nil {
		switch err {
		case subs.ErrNotUpgradeIntent:
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
		err := router.ReaderRepo.ArchiveMember(intent.Result.Snapshot)
		if err != nil {
			sugar.Error(err)
		}

		// Send email
		parcel, err := letter.NewFreeUpgradeParcel(
			account,
			intent.Result.Order,
			intent.Wallet.Sources,
		)

		err = router.Postman.Deliver(parcel)
		if err != nil {
			sugar.Error(err)
			return
		}
	}()

	// Save client app info
	_ = router.postOrderCreation(intent.Result.Order, clientApp)

	_ = render.New(w).NoContent()
}
