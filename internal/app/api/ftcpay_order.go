package api

import (
	"net/http"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
)

// ListOrders loads a list of order under a user.
// Pagination support by adding query parameter:
// page=<int>&per_page=<int>
func (routes FtcPayRoutes) ListOrders(w http.ResponseWriter, req *http.Request) {

	p := gorest.GetPagination(req)
	userIDs := ids.UserIDsFromHeader(req.Header)

	list, err := routes.SubsRepo.ListOrders(userIDs, p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
}

// CMSListOrders list orders of a user tailored for CMS.
// The only difference from ListOrder is that user ids
// is set in query parameter rather than header.
func (routes FtcPayRoutes) CMSListOrders(w http.ResponseWriter, req *http.Request) {
	p := gorest.GetPagination(req)
	userIDs := ids.UserIDsFromQuery(req.Form)

	list, err := routes.SubsRepo.ListOrders(userIDs, p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
}

func (routes FtcPayRoutes) LoadOrder(w http.ResponseWriter, req *http.Request) {
	userIDs := ids.UserIDsFromHeader(req.Header)

	orderID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	order, err := routes.SubsRepo.RetrieveOrder(orderID)
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

func (routes FtcPayRoutes) CMSFindOrder(w http.ResponseWriter, req *http.Request) {

	orderID, err := xhttp.
		GetURLParam(req, "id").
		ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	order, err := routes.SubsRepo.RetrieveOrder(orderID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(order)
}

// RawPaymentResult fetch data from wxpay or alipay order query endpoints and transfer the data as is.
// The response data formats are not always the same one.
func (routes FtcPayRoutes) RawPaymentResult(w http.ResponseWriter, req *http.Request) {
	defer routes.Logger.Sync()
	sugar := routes.Logger.Sugar()

	// Get ftc order id from URL
	orderID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	order, err := routes.SubsRepo.LoadFullOrder(orderID)
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
		client, err := routes.WxPayClients.FindByAppID(order.ID)
		if err != nil {
			_ = render.New(w).InternalServerError(err.Error())
			return
		}

		wxParam, err := client.QueryOrder(wechat.NewOrderQueryParams(order.ID))
		if err != nil {
			_ = render.New(w).InternalServerError(err.Error())
			return
		}
		_ = render.New(w).OK(wxParam)
		return

	case enum.PayMethodAli:
		aliRes, err := routes.AliPayClient.QueryOrder(order.ID)
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
func (routes FtcPayRoutes) VerifyPayment(w http.ResponseWriter, req *http.Request) {
	defer routes.Logger.Sync()
	sugar := routes.Logger.Sugar()

	// Get ftc order id from URL
	orderID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	sugar = sugar.With("orderId", orderID)

	sugar.Info("Start verifying payment")

	order, err := routes.SubsRepo.LoadFullOrder(orderID)
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
	payResult, err := routes.VerifyOrder(order)
	sugar.Infof("Payment result: %+v", payResult)

	if err != nil {
		sugar.Error(err)
		_ = render.New(w).InternalServerError(err.Error())
		return
	}

	go func() {
		err := routes.SubsRepo.SavePayResult(payResult)
		if err != nil {
			sugar.Error(err)
		}
	}()

	if !payResult.IsOrderPaid() {
		sugar.Info("Order is either not paid or already confirmed")

		m, err := routes.ReaderRepo.RetrieveMember(order.CompoundID)
		if err != nil {
			_ = render.New(w).DBError(err)
			return
		}

		_ = render.New(w).OK(ftcpay.ConfirmationResult{
			Payment:    payResult,
			Order:      order,
			Membership: m,
		})

		return
	}

	// If the order is paid, confirm it.
	cfmResult, cfmErr := routes.ConfirmOrder(payResult, order)
	if cfmErr != nil {
		_ = render.New(w).DBError(cfmErr)
		return
	}

	_ = render.New(w).OK(cfmResult)
}
