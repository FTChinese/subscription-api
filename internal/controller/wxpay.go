package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/client"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"net/http"
)

// PlaceWxOrder creates order for wechat pay.
// POST /wxpay/desktop/<tier>/<cycle>?test=true
// 		/wxpay/mobile/<tier>/<cycle>?test=true
//		/wxpay/jsapi/<tier>/<cycle>?test=true
// 		/wxpay/app/<tier>/<cycle>?test=true
// Header: X-User-Id: <uuid>, or X-Union-Id: <wechat union id>, or both.
// Input:
// openID?: string; Required only for payment inside wechat in-house browser.
// TODO: put all those fields in request body
// tier: string; Currently acquired from URL param
// cycle: string; Currently acquired from URL param
func (router PayRouter) PlaceWxOrder(tradeType wechat.TradeType) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		defer router.logger.Sync()
		sugar := router.logger.Sugar()

		sugar.Info("Start placing a wechat order")

		clientApp := client.NewClientApp(req)
		readerIDs := getReaderIDs(req.Header)

		// Find user account.
		account, err := router.readerRepo.FindAccount(readerIDs)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).DBError(err)
			return
		}

		// Parse request body.
		input, err := gatherWxPayInput(tradeType, req)
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

		// Find the client to use for wxpay
		payClient, err := router.wxPayClients.GetClientByPlatform(tradeType)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).InternalServerError(err.Error())
			return
		}

		// Retrieve the plan from DB by edition.
		plan, err := router.prodRepo.PlanByEdition(input.Edition)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).DBError(err)
			return
		}

		sugar.Infof("Selected plan: %+v", plan)

		config := subs.PaymentConfig{
			Account:        account,
			Plan:           plan,
			Method:         enum.PayMethodWx,
			WebhookBaseURL: router.config.WebHookBaseURL(),
			WxAppID:        null.StringFrom(payClient.GetApp().AppID),
		}

		order, err := router.subRepo.CreateOrder(config)
		if err != nil {
			sugar.Error(err)
			router.handleOrderErr(w, err)
			return
		}

		sugar.Infof("Created order: %+v", order)

		go func() {
			_ = router.subRepo.SaveOrderClient(client.OrderClient{
				OrderID: order.ID,
				Client:  clientApp,
			})
		}()

		// Send order to wx
		// UnifiedOrder checks if `return_code` is SUCCESS/FAIL,
		// validate the signature
		// You have to check if return_code == SUCCESS, appid, mch_id, result_code are valid.
		resp, err := payClient.UnifiedOrder(order.WxPay(wechat.UnifiedOrder{
			IP:        clientApp.UserIP.String,
			TradeType: tradeType,
			OpenID:    input.OpenID.String,
		}))

		if err != nil {
			sugar.Error(err)
			_ = render.New(w).BadRequest(err.Error())
			return
		}

		// Convert wxpay's map to struct for easy manipulation.
		uor := wechat.NewUnifiedOrderResp(order.ID, resp)
		go func() {
			_ = router.subRepo.SavePrepayResp(uor)
		}()

		if r := uor.Validate(payClient.GetApp()); r != nil {
			sugar.Info("Invalid unified order response")
			_ = render.New(w).Unprocessable(r)
			return
		}

		switch tradeType {
		// Desktop returns a url that can be turned to QR code
		case wechat.TradeTypeDesktop:
			_ = render.New(w).JSON(http.StatusOK, subs.WxpayBrowserIntent{
				Order:  order,
				QRCode: uor.QRCode.String,
			})

		// Mobile returns a url which is redirect in browser
		case wechat.TradeTypeMobile:
			_ = render.New(w).JSON(http.StatusOK, subs.WxpayBrowserIntent{
				Order:   order,
				MWebURL: uor.MWebURL.String,
			})

		// Create the json data used by js api
		case wechat.TradeTypeJSAPI:
			_ = render.New(w).JSON(http.StatusOK, subs.WxpayEmbedBrowserIntent{
				Order:  order,
				Params: payClient.InWxBrowserParams(uor),
			})

		// Create the json data used by native app.
		case wechat.TradeTypeApp:
			_ = render.New(w).JSON(http.StatusOK, subs.WxpayNativeAppIntent{
				Order:  order,
				Params: payClient.AppParams(uor),
			})
		}
	}
}

// WxWebHook implements 支付结果通知
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_7&index=3
func (router PayRouter) WxWebHook(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	resp := wxpay.Notifies{}

	// Decode Wechat XML request body.
	// If it cannot be decoded, tell wechat to resend it.
	params, err := wechat.DecodeXML(req.Body)
	if err != nil {
		sugar.Error(err)

		if _, err := w.Write([]byte(resp.NotOK(err.Error()))); err != nil {
			sugar.Error(err)
		}

		return
	}

	sugar.Info("Wechat notification decoded")

	// Turn the map to struct
	noti := wechat.NewNotification(params)

	sugar.Infof("Wxpay webhook payload: %+v", noti)

	// Check the status code.
	err = noti.IsStatusValid()
	if err != nil {
		sugar.Error(err)

		if _, err := w.Write([]byte(resp.OK())); err != nil {
			sugar.Error(err)
		}
		return
	}

	// Try to find out which app is in charge of the response.
	payClient, err := router.wxPayClients.GetClientByAppID(noti.AppID.String)

	if err != nil {
		sugar.Error(err)

		if _, err := w.Write([]byte(resp.NotOK(err.Error()))); err != nil {
			sugar.Error(err)
		}

		return
	}

	// Log the response, regardless of whether it is an error
	// or not.
	go func() {
		if err := router.subRepo.SaveWxNotification(noti); err != nil {
			sugar.Error(err)
		}
	}()

	if err := payClient.VerifyNotification(noti); err != nil {
		sugar.Error(err)

		if _, err := w.Write([]byte(resp.OK())); err != nil {
			sugar.Error(err)
		}
		return
	}

	payResult, err := subs.NewWxPayResult(noti)
	if err != nil {
		sugar.Error(err)
		if _, err := w.Write([]byte(resp.OK())); err != nil {
			sugar.Error()
		}

		return
	}

	sugar.Infof("Payment result %+v", payResult)

	confirmed, cfmErr := router.subRepo.ConfirmOrder(payResult)

	// Handle confirmation error.
	if cfmErr != nil {
		sugar.Error(cfmErr)
		go func() {
			_ = router.subRepo.SaveConfirmationResult(
				cfmErr.Schema(payResult.OrderID))
		}()

		if cfmErr.Retry {
			if _, err := w.Write([]byte(resp.NotOK(cfmErr.Error()))); err != nil {
				sugar.Error(err)
			}
		} else {
			if _, err := w.Write([]byte(resp.OK())); err != nil {
				sugar.Error(err)
			}
		}

		return
	}

	go func() {
		router.processCfmResult(confirmed)
	}()

	if _, err := w.Write([]byte(resp.OK())); err != nil {
		sugar.Error(err)
	}
}

// QueryWxOrder implements 查询订单
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
// GET /wxpay/query/{orderId}?app_id=<string>
func (router PayRouter) QueryWxOrder(w http.ResponseWriter, req *http.Request) {

	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Get ftc order id from URL
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

	if order.PaymentMethod != enum.PayMethodWx {
		_ = render.New(w).BadRequest("Order not paid via wechat")
		return
	}

	// The result might not be valid. However we cannot check it for backward compatibility.
	paidResult, respErr := router.queryWxOrder(order)
	if respErr != nil {
		_ = render.New(w).JSON(respErr.StatusCode, respErr)
		return
	}

	_ = render.New(w).OK(paidResult)
}
