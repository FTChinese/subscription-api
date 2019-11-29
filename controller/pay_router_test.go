package controller

import (
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestPayRouter_sendConfirmationEmail(t *testing.T) {
	router := PayRouter{
		env: repository.New(
			test.DB,
			test.Cache,
			util.BuildConfig{},
		),
		postman: test.Postman,
	}

	testRepo := test.NewRepo()

	store := test.NewSubStore(test.MyProfile, reader.AccountKindFtc)

	order := store.MustConfirm(store.MustCreate(test.YearlyStandardLive).ID)
	if err := router.sendConfirmationEmail(order); err != nil {
		t.Error(err)
	}

	order = store.MustConfirm(store.MustCreate(test.YearlyStandardLive).ID)
	if err := router.sendConfirmationEmail(order); err != nil {
		t.Error(err)
	}

	order = store.MustConfirm(store.MustCreate(test.YearlyPremiumLive).ID)
	testRepo.SaveUpgradePlan(store.UpgradePlan)
	testRepo.SaveBalanceSources(store.UpgradePlan.Data)

	if err := router.sendConfirmationEmail(order); err != nil {
		t.Error(err)
	}
}
