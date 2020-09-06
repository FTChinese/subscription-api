package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/client"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/objcoding/wxpay"
	"net/http"
)

// WxPayRouter wraps wxpay and alipay sdk instances.
type WxPayRouter struct {
	clients wechat.PayClients
	PayRouter
}

// NewWxRouter creates a new instance or OrderRouter
func NewWxRouter(baseRouter PayRouter) WxPayRouter {
	r := WxPayRouter{
		clients:   wechat.InitPayClients(),
		PayRouter: baseRouter,
	}

	return r
}

// PlaceOrder creates order for wechat pay.
// POST /wxpay/desktop/<tier>/<cycle>?test=true
// 		/wxpay/mobile/<tier>/<cycle>?test=true
//		/wxpay/jsapi/<tier>/<cycle>?test=true
// 		/wxpay/app/<tier>/<cycle>?test=true
func (router WxPayRouter) PlaceOrder(tradeType wechat.TradeType) http.HandlerFunc {
	sugar := router.logger.Sugar()
	sugar.Infow("Create wxpay order",
		"trace", "WxPayRouter.PlaceOrder",
		"platform", tradeType.String(),
	)

	// Request input:
	// openID?: string; Required only for payment inside wechat in-house browser.
	// TODO: put all those fields in request body
	// tier: string; Currently acquired from URL param
	// cycle: string; Currently acquired from URL param
	// planId: string; Not used yet. In the future we might only use plan id to identify a purchase.
	return func(w http.ResponseWriter, req *http.Request) {
		defer router.logger.Sync()
		sugar.Info("Start placing a wechat order")

		clientApp := client.NewClientApp(req)
		readerIDs := getReaderIDs(req.Header)

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
		payClient, err := router.clients.GetClientByPlatform(tradeType)
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

		isTest := router.isTestAccount(readerIDs, req)

		builder := subs.NewOrderBuilder(readerIDs).
			SetPlan(plan).
			SetPayMethod(enum.PayMethodWx).
			SetWebhookURL(router.config.WebHookBaseURL()).
			SetTest(isTest).
			SetWxAppID(payClient.GetApp().AppID).
			SetWxParams(wechat.UnifiedOrder{
				IP:        clientApp.UserIP.String,
				TradeType: tradeType,
				OpenID:    input.OpenID.String,
			})

		order, err := router.subRepo.CreateOrder(builder)
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
		resp, err := payClient.UnifiedOrder(builder.WxpayParams())
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
			_ = render.New(w).JSON(http.StatusOK, subs.WxpayBrowserOrder{
				Order:  order,
				QRCode: uor.QRCode.String,
			})

		// Mobile returns a url which is redirect in browser
		case wechat.TradeTypeMobile:
			_ = render.New(w).JSON(http.StatusOK, subs.WxpayBrowserOrder{
				Order:   order,
				MWebURL: uor.MWebURL.String,
			})

		// Create the json data used by js api
		case wechat.TradeTypeJSAPI:
			_ = render.New(w).JSON(http.StatusOK, subs.WxpayEmbedBrowserOrder{
				Order:  order,
				Params: payClient.InWxBrowserParams(uor),
			})

		// Create the json data used by native app.
		case wechat.TradeTypeApp:
			_ = render.New(w).JSON(http.StatusOK, subs.WxpayNativeAppOrder{
				Order:  order,
				Params: payClient.AppParams(uor),
			})
		}
	}
}

// WebHook implements 支付结果通知
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_7&index=3
func (router WxPayRouter) WebHook(w http.ResponseWriter, req *http.Request) {
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
	payClient, err := router.clients.GetClientByAppID(noti.AppID.String)

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

	payResult, err := subs.NewPaymentResultWx(noti)
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

	if !confirmed.Snapshot.IsZero() {
		_ = router.readerRepo.BackUpMember(confirmed.Snapshot)
	}

	go func() {
		if err := router.sendConfirmationEmail(confirmed.Order); err != nil {
			sugar.Error(err)
		}
	}()

	if _, err := w.Write([]byte(resp.OK())); err != nil {
		sugar.Error(err)
	}
}

// OrderQuery implements 查询订单
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
// GET /wxpay/query/{orderId}?app_id=<string>
func (router WxPayRouter) OrderQuery(w http.ResponseWriter, req *http.Request) {

	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Get ftc order id from URL
	orderID, err := getURLParam(req, "orderId").ToString()

	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	appID := getWxAppID(req)
	// Get the client to query wechat api.
	payClient, err := router.clients.GetClientByAppID(appID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	reqParams := make(wxpay.Params)
	reqParams.SetString("out_trade_no", orderID)

	// Send query to Wechat server
	// Returns the parsed response as a map.
	// It checks if the response contains `return_code` key.
	// If return_code == FAIL, it does not returns error.
	// If return_code == SUCCESS, it verifies the signature.
	respParams, err := payClient.OrderQuery(reqParams)

	// If there are any errors when querying order.
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).InternalServerError(err.Error())
		return
	}

	sugar.Infof("Wechat order found")

	// Response:
	// {message: "", {field: status, code: fail} }
	// {message: "", {field: result, code: "ORDERNOTEXIST" | "SYSTEMERROR"} }
	resp := wechat.NewOrderQueryResp(respParams)
	go func() {
		if err := router.subRepo.SaveWxQueryResp(resp); err != nil {
			sugar.Error(err)
		}
	}()

	if r := resp.Validate(payClient.GetApp()); r != nil {
		sugar.Info("Response invalid")

		if r.Field == "result" && r.Code == "ORDERNOTEXIST" {
			_ = render.New(w).NotFound()
			return
		}

		_ = render.New(w).Unprocessable(r)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, resp.ToQueryResult())
}
