package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"net/http"
)

// ListOrders loads a list of membership change history.
// Pagination support by adding query parameter:
// page=<int>&per_page=<int>
func (router SubsRouter) ListOrders(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	p := gorest.GetPagination(req)
	userIDs := getReaderIDs(req.Header)

	list, err := router.SubsRepo.ListOrders(userIDs, p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
}

func (router SubsRouter) LoadOrder(w http.ResponseWriter, req *http.Request) {
	userIDs := getReaderIDs(req.Header)

	orderID, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	order, err := router.SubsRepo.RetrieveOrder(orderID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if order.CompoundID != userIDs.CompoundID {
		_ = render.New(w).NotFound("")
		return
	}

	_ = render.New(w).OK(order)
}

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

	// Start fetching payment result from Ali/Wx
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

		m, err := router.SubsRepo.RetrieveMember(order.CompoundID)
		if err != nil {
			_ = render.New(w).DBError(err)
			return
		}

		_ = render.New(w).OK(subs.ConfirmationResult{
			Payment:    payResult,
			Order:      order,
			Membership: m,
			Snapshot:   reader.MemberSnapshot{},
		})

		return
	}

	// If the order is paid, confirm it.
	cfmResult, cfmErr := router.ConfirmOrder(payResult, order)
	if cfmErr != nil {
		_ = render.New(w).DBError(cfmErr)
		return
	}

	_ = render.New(w).OK(cfmResult)
}
