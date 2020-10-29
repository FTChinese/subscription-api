package controller

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"net/http"
)

func (router PayRouter) ManualConfirm(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Get ftc order id from URL
	orderID, err := getURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error()
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	order, err := router.subRepo.RetrieveOrder(orderID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if order.IsConfirmed() {
		_ = render.New(w).Forbidden("Order already confirmed")
		return
	}

	var paidResult subs.PaymentResult
	switch order.PaymentMethod {
	case enum.PayMethodWx:
		paidResult, err = router.verifyWxPayment(order)

	case enum.PayMethodAli:
		paidResult, err = router.verifyAliPayment(order)

	default:
		sugar.Error("Manual confirmation: not ali or wx order")
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: "Only orders paid via ali or wx is allowed",
			Field:   "payMethod",
			Code:    render.CodeInvalid,
		})
		return
	}

	if err != nil {
		sugar.Error(err)
		_ = render.New(w).InternalServerError(err.Error())
		return
	}

	sugar.Infof("Payment result: %+v", paidResult)

	paidResult.ConfirmedAt = chrono.TimeNow()

	confirmed, cfmErr := router.subRepo.ConfirmOrder(paidResult)
	if cfmErr != nil {
		sugar.Error(cfmErr)
		_ = render.New(w).DBError(cfmErr)
		return
	}

	if !confirmed.Snapshot.IsZero() {
		go func() {
			_ = router.readerRepo.BackUpMember(confirmed.Snapshot)
		}()
	}

	go func() {
		if err := router.sendConfirmationEmail(confirmed.Order); err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(confirmed.Order)
}
