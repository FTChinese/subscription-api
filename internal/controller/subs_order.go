package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"net/http"
)

// VerifyPayment checks against payment provider's api to get
// the payment result of an order.
// GET /wxpay/query/{orderId}
// GET /alipay/query/{orderId}
// POST /orders/{id}/verify-payment
func (router SubsRouter) VerifyPayment(verifyOnly bool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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

		if verifyOnly {
			_ = render.New(w).OK(payResult)
			return
		}

		if !payResult.IsOrderPaid() || order.IsConfirmed() {
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
		if cfmErr != nil {
			_ = render.New(w).DBError(err)
			return
		}

		_ = render.New(w).OK(subs.VerificationResult{
			PaymentResult:      payResult,
			ConfirmationResult: cfmResult,
		})
	}
}
