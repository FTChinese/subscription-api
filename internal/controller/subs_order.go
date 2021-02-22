package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"net/http"
)

// RawPaymentResult fetch data from wxpay or alipay order query endpoints and transfer the data as is.
// The response data formats are not always the same one.
func (router SubsRouter) RawPaymentResult(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	// Get ftc order id from URL
	orderID, err := getURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	order, err := router.SubsRepo.LoadFullOrder(orderID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if !order.IsAliWxPay() {
		sugar.Info("Not ali or wx pay")
		_ = render.New(w).BadRequest("Order not paid via alipay or wxpay")
		return
	}

	switch order.PaymentMethod {
	case enum.PayMethodWx:
		wxParam, err := router.WxPayClients.QueryOrderRaw(order)
		if err != nil {
			_ = render.New(w).InternalServerError(err.Error())
			return
		}
		_ = render.New(w).OK(wxParam)
		return

	case enum.PayMethodAli:
		aliRes, err := router.AliPayClient.QueryOrder(order.ID)
		if err != nil {
			_ = render.New(w).InternalServerError(err.Error())
			return
		}
		_ = render.New(w).OK(aliRes)
		return
	}

	_ = render.New(w).BadRequest("Unknown payment method")
}

// VerifyPayment checks against payment provider's api to get
// the payment result of an order.
// GET /wxpay/query/{orderId}
// GET /alipay/query/{orderId}
// POST /orders/{id}/verify-payment
func (router SubsRouter) VerifyPayment(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	// Get ftc order id from URL
	orderID, err := getURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	sugar = sugar.With("orderId", orderID)

	sugar.Info("Start verifying payment")

	order, err := router.SubsRepo.LoadFullOrder(orderID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if !order.IsAliWxPay() {
		sugar.Info("Not ali or wx pay")
		_ = render.New(w).BadRequest("Order not paid via alipay or wxpay")
		return
	}

	payResult, err := router.VerifyOrder(order)
	sugar.Infof("Payment result: %+v", payResult)

	if err != nil {
		sugar.Error(err)
		_ = render.New(w).InternalServerError(err.Error())
		return
	}

	go func() {
		err := router.SubsRepo.SavePayResult(payResult)
		if err != nil {
			sugar.Error(err)
		}
	}()

	if !payResult.IsOrderPaid() {
		sugar.Info("Order is either not paid or already confirmed")

		m, err := router.ReaderRepo.RetrieveMember(order.MemberID)
		if err != nil {
			_ = render.New(w).DBError(err)
			return
		}

		_ = render.New(w).OK(subs.VerificationResult{
			PaymentResult: payResult,
			ConfirmationResult: subs.ConfirmationResult{
				Order:      order,
				Membership: m,
				Payment:    payResult,
			},
		})

		return
	}

	cfmResult, cfmErr := router.ConfirmOrder(payResult, order)
	if cfmErr != nil && cfmErr.Retry {
		_ = render.New(w).DBError(cfmErr)
		return
	}

	_ = render.New(w).OK(subs.VerificationResult{
		PaymentResult:      payResult,
		ConfirmationResult: cfmResult,
	})
}
