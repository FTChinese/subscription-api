package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/go-rest/view"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/client"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"net/http"
)

const (
	success = "success"
	fail    = "fail"
)

// PlaceOrder creates an http handler function depending
// on the device platform.
//
// 	POST /<desktop|mobile|app>/{tier}/{cycle}?<test=true>&<return_url=xxx>
//
// Header: X-User-Id or X-Union-Id, or both.
// `return_url` parameter is only required for apps running on ftacademy.cn. It is deprecated and changed to request body.
//
// Input:
// returnUrl?: string; Only for browsers.
// tier: string;
// cycle: string;
// planId?: string;
func (router PayRouter) PlaceAliOrder(kind ali.EntryKind) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer router.logger.Sync()
		sugar := router.logger.Sugar()

		err := req.ParseForm()
		if err != nil {
			sugar.Error(err)
			_ = view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		clientApp := client.NewClientApp(req)
		readerIDs := getReaderIDs(req.Header)

		// Find user account.
		account, err := router.readerRepo.FindAccount(readerIDs)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).DBError(err)
			return
		}

		input, err := gatherAliPayInput(req)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).BadRequest(err.Error())
			return
		}
		if ve := input.Validate(); ve != nil {
			sugar.Error(err)
			_ = render.New(w).Unprocessable(ve)
			return
		}

		// Find pricing plan.
		plan, err := router.prodRepo.PlanByEdition(input.Edition)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).BadRequest(err.Error())
			return
		}

		config := subs.PaymentConfig{
			Account:        account,
			Plan:           plan,
			Method:         enum.PayMethodAli,
			WebhookBaseURL: router.config.WebHookBaseURL(),
		}

		order, err := router.subRepo.CreateOrder(config)
		if err != nil {
			_ = render.New(w).InternalServerError(err.Error())
			return
		}

		sugar.Infof("Created order: %+v", order)

		go func() {
			err := router.subRepo.SaveOrderClient(client.OrderClient{
				OrderID: order.ID,
				Client:  clientApp,
			})
			if err != nil {
				sugar.Error(err)
			}
		}()

		switch kind {
		case ali.EntryApp:
			// Generate signature and creates a string of query
			// parameter.
			queryStr, err := router.aliPay.TradeAppPay(
				order.AliAppPay())

			sugar.Infof("App pay param: %+v\n", queryStr)

			if err != nil {
				sugar.Error(err)
				_ = render.New(w).BadRequest(err.Error())
				return
			}

			_ = render.New(w).
				JSON(http.StatusOK, subs.AlipayNativeIntent{
					Order: order,
					Param: queryStr,
				})
			return

		case ali.EntryDesktopWeb:
			redirectURL, err := router.aliPay.TradePagePay(
				order.AliDesktopPay(input.ReturnURL),
			)
			if err != nil {
				sugar.Error(err)
				_ = render.New(w).BadRequest(err.Error())
				return
			}

			sugar.Infof("Ali desktop browser redirect url: %+v\n", redirectURL)

			_ = render.New(w).JSON(http.StatusOK, subs.AlipayBrowserIntent{
				Order:       order,
				RedirectURL: redirectURL.String(),
			})

		case ali.EntryMobileWeb:
			redirectURL, err := router.aliPay.TradeWapPay(
				order.AliWapPay(input.ReturnURL),
			)
			if err != nil {
				sugar.Error(err)
				_ = render.New(w).BadRequest(err.Error())
				return
			}
			sugar.Infof("Ali mobile browser redirect url: %+v\n", redirectURL)

			_ = render.New(w).JSON(http.StatusOK, subs.AlipayBrowserIntent{
				Order:       order,
				RedirectURL: input.ReturnURL,
			})
		}
	}
}

// Query verifies the payment status of an order against alipay api.
// GET /alipay/query/{orderId}
func (router PayRouter) QueryAliOrder(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	orderID, err := getURLParam(req, "orderId").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	order, err := router.subRepo.RetrieveOrder(orderID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if order.PaymentMethod != enum.PayMethodAli {
		_ = render.New(w).BadRequest("Order not paid via alipay")
		return
	}

	result, respErr := router.queryAliOrder(order)
	if respErr != nil {
		_ = render.New(w).JSON(respErr.StatusCode, respErr)
		return
	}

	_ = render.New(w).OK(result)
}

// AliWebHook handles alipay server-side notification.
// POST /webhook/alipay
func (router PayRouter) AliWebHook(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	err := req.ParseForm()
	if err != nil {
		sugar.Error(err)

		if _, err := w.Write([]byte(fail)); err != nil {
			sugar.Error(err)
		}
		return
	}

	// If err is nil, then the signature is verified.
	noti, err := router.aliPay.GetTradeNotification(req)
	sugar.Infof("+%v", noti)

	if err != nil {
		sugar.Error(err)

		if _, err := w.Write([]byte(fail)); err != nil {
			sugar.Error(err)
		}
		return
	}

	// 4、验证app_id是否为该商户本身
	if noti.AppId != router.aliAppID {
		sugar.Infof("Expected %s, actual %s", router.aliAppID, noti.AppId)

		if _, err := w.Write([]byte(fail)); err != nil {
			sugar.Error(err)
		}
		return
	}

	go func() {
		if err := router.subRepo.SaveAliNotification(*noti); err != nil {
			sugar.Error(err)
		}
	}()

	// 在支付宝的业务通知中，只有交易通知状态为TRADE_SUCCESS或TRADE_FINISHED时，支付宝才会认定为买家付款成功。
	if !ali.IsPaySuccess(noti) {
		sugar.Infof("Status %s", noti.TradeStatus)

		if ali.ShouldRetry(noti) {
			if _, err := w.Write([]byte(fail)); err != nil {
				sugar.Error(err)
			}
			return
		}

		if _, err := w.Write([]byte(success)); err != nil {
			sugar.Error(err)
		}
		return
	}

	payResult, err := subs.NewAliPayResult(noti)

	// 1、商户需要验证该通知数据中的out_trade_no是否为商户系统中创建的订单号
	// 2、判断total_amount是否确实为该订单的实际金额（即商户订单创建时的金额）
	confirmed, cfmErr := router.subRepo.ConfirmOrder(payResult)

	if cfmErr != nil {
		sugar.Error(cfmErr)
		go func() {
			_ = router.subRepo.SaveConfirmationResult(
				cfmErr.Schema(payResult.OrderID),
			)
		}()

		if cfmErr.Retry {
			if _, err := w.Write([]byte(fail)); err != nil {
				sugar.Error(err)
			}
			return
		} else {
			if _, err := w.Write([]byte(success)); err != nil {
				sugar.Error(err)
			}
			return
		}
	}

	go func() {
		router.processCfmResult(confirmed)
	}()

	if _, err := w.Write([]byte(success)); err != nil {
		sugar.Error(err)
	}
}
