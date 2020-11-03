package controller

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"net/http"
)

func (router SubsRouter) ManualConfirm(w http.ResponseWriter, req *http.Request) {
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

	// Return error or data?
	if order.IsConfirmed() {
		_ = render.New(w).Forbidden("Order already confirmed")
		return
	}

	if !order.IsAliWxPay() {
		_ = render.New(w).BadRequest("order not paid via ali or wx")
		return
	}

	var payResult subs.PaymentResult
	switch order.PaymentMethod {
	case enum.PayMethodWx:
		payResult, err = router.verifyWxPayment(order)

	case enum.PayMethodAli:
		payResult, err = router.verifyAliPayment(order)
	}

	sugar.Infof("Payment result: %+v", payResult)

	if err != nil {
		sugar.Error(err)
		_ = render.New(w).InternalServerError(err.Error())
		return
	}

	if !payResult.IsOrderPaid() {
		_ = render.New(w).BadRequest("This order is not paid")
		return
	}

	payResult.ConfirmedAt = chrono.TimeNow()

	confirmed, err := router.processPaymentResult(payResult)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(confirmed.Order)
}

// VerifyPayment checks against payment provider's api to get
// the payment result of an order.
// GET /wxpay/query/{orderId}
// GET /alipay/query/{orderId}
// POST /orders/{id}/verify-payment
func (router SubsRouter) VerifyPayment(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Get ftc order id from URL
	orderID, err := getURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	sugar.Infow("Verify payment",
		"orderId", orderID)

	order, err := router.subRepo.RetrieveOrder(orderID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if !order.IsAliWxPay() {
		_ = render.New(w).BadRequest("Order not paid vai alipay or wxpay")
		return
	}

	var payResult subs.PaymentResult
	switch order.PaymentMethod {
	case enum.PayMethodWx:
		payResult, err = router.verifyWxPayment(order)

	case enum.PayMethodAli:
		payResult, err = router.verifyAliPayment(order)
	}

	if err != nil {
		sugar.Error(err)
		_ = render.New(w).InternalServerError(err.Error())
		return
	}

	sugar.Infof("Payment result: %+v", payResult)

	if !payResult.IsOrderPaid() {
		_ = render.New(w).OK(payResult)
		return
	}

	if order.IsConfirmed() {
		_ = render.New(w).OK(payResult)
		return
	}

	_, _ = router.processPaymentResult(payResult)

	_ = render.New(w).OK(payResult)

}
