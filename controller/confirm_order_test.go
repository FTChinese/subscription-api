package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository/subrepo"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
	"time"
)

func mockOrderAndConfirm(id reader.MemberID) subscription.Order {
	router := PayRouter{
		subEnv: subrepo.New(test.DB, test.Cache, util.BuildConfig{
			Sandbox:    false,
			Production: false,
		}),
		postman: test.Postman,
	}

	order, err := router.createOrder(
		id,
		test.YearlyStandardLive,
		enum.PayMethodWx,
		test.RandomClientApp(),
		null.StringFrom(test.WxPayApp.AppID),
	)

	if err != nil {
		panic(err)
	}

	paymentResult := subscription.PaymentResult{
		Amount:      int64(test.YearlyStandardLive.NetPrice * 100),
		OrderID:     order.ID,
		ConfirmedAt: time.Now(),
	}

	confirmedOrder, resultErr := router.confirmPayment(paymentResult)
	if resultErr != nil {
		panic(err)
	}

	return confirmedOrder
}

func TestPayRouter_confirmPayment(t *testing.T) {
	profile := test.NewProfile()
	accountID := profile.AccountID(reader.AccountKindFtc)

	orderToCreate := mockOrderAndConfirm(accountID)
	t.Logf("Creation order: %+v", orderToCreate)

	orderToRenew := mockOrderAndConfirm(accountID)
	t.Logf("Renewal order: %+v", orderToRenew)
}
