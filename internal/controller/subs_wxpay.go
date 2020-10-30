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
func (router SubsRouter) PlaceWxOrder(tradeType wechat.TradeType) http.HandlerFunc {

	// Find the client to use for wxpay
	payClient, err := router.wxPayClients.ClientByPlatform(tradeType)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		defer router.logger.Sync()
		sugar := router.logger.Sugar()

		sugar.Info("Start placing a wechat pi")

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

		// Retrieve the plan from DB by edition.
		plan, err := router.prodRepo.PlanByEdition(input.Edition)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).DBError(err)
			return
		}

		sugar.Infof("Selected plan: %+v", plan)

		config := subs.NewPayment(account, plan).
			WithWxpay(payClient.GetApp(), router.config.WebHookBaseURL())

		pi, err := router.subRepo.CreateOrder(config)
		if err != nil {
			sugar.Error(err)
			router.handleOrderErr(w, err)
			return
		}

		sugar.Infof("Created order: %+v", pi.Order)

		go func() {
			if pi.Kind == enum.OrderKindUpgrade {
				err := router.subRepo.SaveProratedOrders(pi.ProratedOrders())
				if err != nil {
					sugar.Error(err)
				}
			}

			err := router.subRepo.LogOrderMeta(subs.OrderMeta{
				OrderID: pi.Order.ID,
				Client:  clientApp,
			})
			if err != nil {
				sugar.Error(err)
			}
		}()

		// Send order to wx
		// UnifiedOrder checks if `return_code` is SUCCESS/FAIL,
		// validate the signature
		// You have to check if return_code == SUCCESS, appid, mch_id, result_code are valid.
		wxOrder, err := payClient.CreateOrder(wechat.OrderReq{
			Body:          pi.Item.Plan.PaymentTitle(pi.Kind),
			SellerOrderID: pi.Order.ID,
			TotalAmount:   pi.Order.AmountInCent(),
			WebhookURL:    pi.WebhookURL,
			ProductID:     pi.Item.Plan.ID,
			TxKind:        tradeType,
			UserIP:        clientApp.UserIP.String,
			OpenID:        input.OpenID,
		})

		// Save raw response.
		go func() {
			err := router.subRepo.SavePrepayResp(wxOrder)
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
			_ = render.New(w).OK(pi.WxPayDesktopIntent(wxOrder))

		// Mobile returns a url which is redirect in browser
		case wechat.TradeTypeMobile:
			_ = render.New(w).OK(pi.WxPayMobileIntent(wxOrder))

		// Create the json data used by js api
		case wechat.TradeTypeJSAPI:
			_ = render.New(w).OK(
				pi.WxPayJSApiIntent(
					payClient.SignJSApiParams(wxOrder),
				),
			)

		// Create the json data used by native app.
		case wechat.TradeTypeApp:
			_ = render.New(w).OK(
				pi.WxNativeAppIntent(
					payClient.SignAppParams(wxOrder),
				),
			)
		}
	}
}

// WxWebHook implements 支付结果通知
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_7&index=3
func (router SubsRouter) WxWebHook(w http.ResponseWriter, req *http.Request) {
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
	sugar.Infof("Wxpay webhook payload: %+v", params)

	// Turn the map to struct
	noti := wechat.NewNotification(params)
	// Log the response, regardless of whether it is an error
	// or not.
	go func() {
		if err := router.subRepo.SaveWxNotification(noti); err != nil {
			sugar.Error(err)
		}
	}()

	// Try to find out which app is in charge of the response.
	payClient, err := router.wxPayClients.ClientByAppID(noti.AppID.String)

	if err != nil {
		sugar.Error(err)

		if _, err := w.Write([]byte(resp.NotOK(err.Error()))); err != nil {
			sugar.Error(err)
		}

		return
	}

	if err := payClient.ValidateWebhook(noti); err != nil {
		sugar.Error(err)

		if _, err := w.Write([]byte(resp.OK())); err != nil {
			sugar.Error(err)
		}
		return
	}

	payResult, err := subs.NewWxWebhookResult(noti)
	if err != nil {
		sugar.Error(err)
		if _, err := w.Write([]byte(resp.OK())); err != nil {
			sugar.Error()
		}

		return
	}

	sugar.Infof("Payment result %+v", payResult)

	// TODO: refactor confirmation
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
