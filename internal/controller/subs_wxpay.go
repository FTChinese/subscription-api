package controller

import (
	"errors"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"net/http"
)

// WxPay creates order for wechat pay.
// Header: X-User-Id: <uuid>, or X-Union-Id: <wechat union id>, or both.
// Input:
// openID?: string; Required only for payment inside wechat in-house browser.
// tier: string; Currently acquired from URL param
// cycle: string; Currently acquired from URL param
// planId: string;
func (router SubsRouter) WxPay(tradeType wechat.TradeType) http.HandlerFunc {

	webhookURL := subs.WebhookURL(router.config.Sandbox(), enum.PayMethodWx)

	// Find the client to use for wxpay
	payClient, err := router.WxPayClients.ClientByPlatform(tradeType)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		defer router.Logger.Sync()
		sugar := router.Logger.Sugar()

		sugar.Info("Start creating a wechat order")

		clientApp := footprint.NewClient(req)
		readerIDs := getReaderIDs(req.Header)

		// Find user account.
		acnt, err := router.SubsRepo.FindBaseAccount(readerIDs)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).DBError(err)
			return
		}

		// Parse request body.
		input := ftcpay.NewWxPayReq(tradeType)
		if err := gorest.ParseJSON(req.Body, &input); err != nil {
			sugar.Error(err)
			_ = render.New(w).BadRequest(err.Error())
			return
		}
		if ve := input.Validate(); ve != nil {
			sugar.Error(err)
			_ = render.New(w).Unprocessable(ve)
			return
		}

		item, re := router.loadCheckoutItem(input.OrderParams, !acnt.IsTest())
		if re != nil {
			sugar.Error(re)
			_ = render.New(w).JSON(re.StatusCode, re)
			return
		}

		counter := ftcpay.Counter{
			BaseAccount:  acnt,
			CheckoutItem: item,
			PayMethod:    enum.PayMethodWx,
			WxAppID:      null.StringFrom(payClient.GetApp().AppID),
		}

		pi, err := router.SubsRepo.CreateOrder(counter)
		if err != nil {
			sugar.Error(err)
			router.handleOrderErr(w, err)
			return
		}

		sugar.Infof("Created order: %+v", pi.Order)

		err = router.postOrderCreation(pi.Order, clientApp)
		if err != nil {
			_ = render.New(w).DBError(err)
			return
		}

		// Send order to wx
		// UnifiedOrder checks if `return_code` is SUCCESS/FAIL,
		// validate the signature
		// You have to check if return_code == SUCCESS, appid, mch_id, result_code are valid.
		wxOrder, err := payClient.CreateOrder(wechat.OrderReq{
			Body:          subs.PaymentTitle(pi.Order.Kind, pi.Order.Edition),
			SellerOrderID: pi.Order.ID,
			TotalAmount:   pi.Order.AmountInCent(),
			WebhookURL:    webhookURL,
			ProductID:     pi.Order.PlanID,
			TxKind:        tradeType,
			UserIP:        clientApp.UserIP.String,
			OpenID:        input.OpenID,
		})

		// Save raw response.
		go func() {
			err := router.SubsRepo.SavePrepayResp(wxOrder)
			if err != nil {
				sugar.Error(err)
			}
		}()

		if err != nil {
			sugar.Error(err)
			_ = render.New(w).BadRequest(err.Error())
			return
		}

		// Validate wechat response.
		err = wxOrder.Validate(payClient.GetApp())
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).InternalServerError(err.Error())
			return
		}

		switch tradeType {
		// Desktop returns a url that can be turned to QR code
		case wechat.TradeTypeDesktop:
			_ = render.New(w).OK(subs.NewWxPayDesktopIntent(pi, wxOrder))

		// Mobile returns a url which is redirect in browser
		case wechat.TradeTypeMobile:
			_ = render.New(w).OK(subs.NewWxPayMobileIntent(pi, wxOrder))

		// Create the json data used by js api
		case wechat.TradeTypeJSAPI:
			_ = render.New(w).OK(
				subs.NewWxPayJSApiIntent(
					pi,
					payClient.SignJSApiParams(wxOrder),
				),
			)

		// Create the json data used by native app.
		case wechat.TradeTypeApp:
			_ = render.New(w).OK(
				subs.NewWxNativeAppIntent(
					pi,
					payClient.SignAppParams(wxOrder),
				),
			)
		}
	}
}

// WxWebHook implements 支付结果通知
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_7&index=3
func (router SubsRouter) WxWebHook(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	sugar.Info("Wxpay webhook received message")

	resp := wxpay.Notifies{}

	var send = func(err error) {
		var e error
		if err != nil {
			_, e = w.Write([]byte(resp.NotOK(err.Error())))
		} else {
			_, e = w.Write([]byte(resp.OK()))
		}

		if e != nil {
			sugar.Error(e)
		}
	}

	// Decode Wechat XML request body.
	// If it cannot be decoded, tell wechat to resend it.
	sugar.Info("Getting webhook payload")
	payload, err := router.WxPayClients.GetWebhookPayload(req)
	//params, err := wechat.DecodeXML(req.Body)
	if err != nil {
		sugar.Error(err)
		send(err)
		return
	}

	// Turn the map to struct
	//payload := wechat.NewNotification(params)
	// Log the response, regardless of whether it is an error
	// or not.
	go func() {
		sugar.Info("Saving wxpay webhook raw payload")
		if err := router.SubsRepo.SaveWxNotification(payload); err != nil {
			sugar.Error(err)
		}
	}()

	sugar.Info("Checking payload unprocessable")
	if payload.IsUnprocessable() {
		sugar.Error(payload.UnprocessableMsg())
		send(errors.New(payload.ErrorMessage.String))
		return
	}

	// Should check this?
	//if err := payload.EnsureSuccess(); err != nil {
	//	sugar.Error(err)
	//	_, _ = w.Write([]byte(resp.OK()))
	//	return
	//}

	payResult := subs.NewWxWebhookResult(payload)

	sugar.Info("Start processing wx webhook")
	_, cfmErr := router.processWebhookResult(payResult)

	// Handle confirmation error.
	if cfmErr != nil {
		sugar.Error(cfmErr)
		if cfmErr.Retry {
			send(cfmErr)
		} else {
			send(nil)
		}

		return
	}

	send(nil)
}
