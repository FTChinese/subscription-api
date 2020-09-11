package controller

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"net/http"
)

func (router PayRouter) ManualConfirm(w http.ResponseWriter, req *http.Request) {
	defer logger.Sync()
	sugar := logger.Sugar()

	// Get ftc order id from URL
	orderID, err := getURLParam(req, "orderId").ToString()
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

	var paidResult subs.PaymentResult
	var respErr *render.ResponseError
	switch order.PaymentMethod {
	case enum.PayMethodWx:
		paidResult, respErr = router.queryWxOrder(order)

	case enum.PayMethodAli:
		paidResult, respErr = router.queryAliOrder(order)

	default:
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: "Only orders paid via ali or wx is allowed",
			Field:   "payMethod",
			Code:    render.CodeInvalid,
		})
		return
	}

	if respErr != nil {
		_ = render.New(w).JSON(respErr.StatusCode, respErr)
		return
	}

	paidResult.ConfirmedAt = chrono.TimeNow()

	confirmed, cfmErr := router.subRepo.ConfirmOrder(paidResult)
	if cfmErr != nil {
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
