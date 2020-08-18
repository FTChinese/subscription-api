package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/models/subscription"
	builder2 "github.com/FTChinese/subscription-api/pkg/builder"
	"github.com/FTChinese/subscription-api/pkg/client"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/objcoding/wxpay"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
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
func (router WxPayRouter) PlaceOrder(tradeType wechat.TradeType) http.HandlerFunc {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	sugar.Infow("Create wxpay order",
		"trace", "WxPayRouter.PlaceOrder",
		"platform", tradeType.String(),
	)

	// Request input:
	// OpenID?: string;
	// TODO: put all those fields in request body
	// tier: string;
	// cycle: string;
	// planId: string;
	// ftcId: string;
	// unionId: string;
	return func(w http.ResponseWriter, req *http.Request) {
		defer logger.Sync()
		sugar.Info("Start placing a wechat order")

		clientApp := client.NewClientApp(req)

		sugar.Infof("Client app: %+v", clientApp)

		// Parse request body.
		input := wechat.NewPayInput(tradeType)
		if err := gorest.ParseJSON(req.Body, &input); err != nil {
			_ = render.New(w).BadRequest(err.Error())
			return
		}
		if ve := input.Validate(); ve != nil {
			_ = render.New(w).Unprocessable(ve)
			return
		}

		edition, err := GetEdition(req)
		if err != nil {
			_ = render.New(w).BadRequest(err.Error())
			return
		}

		// Find the client to use for wxpay
		payClient, err := router.clients.GetClientByPlatform(tradeType)
		if err != nil {
			sugar.Error(err.Error())
			_ = render.New(w).InternalServerError(err.Error())
			return
		}

		// Get ftc user id or wechat union id.
		userID, _ := GetUserID(req.Header)

		expPlan, err := router.prodRepo.PlanByEdition(edition)
		if err != nil {
			_ = render.New(w).DBError(err)
			return
		}

		builder := builder2.NewOrderBuilder(userID).
			SetPlan(expPlan).
			SetPayMethod(enum.PayMethodWx).
			SetWxAppID(payClient.GetApp().AppID).
			SetClient(clientApp).
			SetWxParams(wechat.UnifiedOrder{
				TradeType: tradeType,
				OpenID:    input.OpenID.String,
			}).
			SetEnvironment(router.config.Live())

		order, err := router.subEnv.CreateOrder(builder)

		if err != nil {
			sugar.Error(err)
			router.handleOrderErr(w, err)
			return
		}

		sugar.Infof("Created order: %+v", order)

		go func() {
			_ = router.subEnv.SaveOrderClient(builder.ClientApp())
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
			_ = router.subEnv.SavePrepayResp(uor)
		}()

		if r := uor.Validate(payClient.GetApp()); r != nil {
			sugar.Info("Invalid unified order response")
			_ = render.New(w).Unprocessable(r)
			return
		}

		switch tradeType {
		// Desktop returns a url that can be turned to QR code
		case wechat.TradeTypeDesktop:
			_ = render.New(w).JSON(http.StatusOK, subscription.WxpayBrowserOrder{
				Order:  order,
				QRCode: uor.QRCode.String,
			})

		// Mobile returns a url which is redirect in browser
		case wechat.TradeTypeMobile:
			_ = render.New(w).JSON(http.StatusOK, subscription.WxpayBrowserOrder{
				Order:   order,
				MWebURL: uor.MWebURL.String,
			})

		// Create the json data used by js api
		case wechat.TradeTypeJSAPI:
			_ = render.New(w).JSON(http.StatusOK, subscription.WxpayEmbedBrowserOrder{
				Order:  order,
				Params: payClient.InWxBrowserParams(uor),
			})

		// Create the json data used by native app.
		case wechat.TradeTypeApp:
			params := payClient.AppParams(uor)
			_ = render.New(w).JSON(http.StatusOK, subscription.WxpayNativeAppOrder{
				Order:          order,
				AppOrderParams: params,
				Params:         payClient.AppParams(uor),
			})
		}
	}
}

// WebHook implements 支付结果通知
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_7&index=3
func (router WxPayRouter) WebHook(w http.ResponseWriter, req *http.Request) {

	logger := logrus.WithFields(logrus.Fields{
		"trace": "WxPayRouter.WebHook()",
	})
	resp := wxpay.Notifies{}

	// Decode Wechat XML request body.
	// If it cannot be decoded, tell wechat to resend it.
	params, err := wechat.DecodeXML(req.Body)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"event": "DecodeXML",
		}).Error(err)

		if _, err := w.Write([]byte(resp.NotOK(err.Error()))); err != nil {
			logger.Error(err)
		}

		return
	}

	logger.WithFields(logrus.Fields{
		"param": params,
	}).Info("Wechat notification decoded")

	// Turn the map to struct
	noti := wechat.NewNotification(params)

	// Check the status code.
	err = noti.IsStatusValid()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"event":   "InvalidStatus",
			"orderId": noti.FTCOrderID,
		}).Error(err)

		if _, err := w.Write([]byte(resp.OK())); err != nil {
			logger.Error(err)
		}
		return
	}

	// Try to find out which app is in charge of the response.
	payClient, err := router.clients.GetClientByAppID(noti.AppID.String)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"event":   "FindWechatClient",
			"orderId": noti.FTCOrderID,
		}).Error(err)

		if _, err := w.Write([]byte(resp.NotOK(err.Error()))); err != nil {
			logger.Error(err)
		}

		return
	}

	// Log the response, regardless of whether it is an error
	// or not.
	go func() {
		if err := router.subEnv.SaveWxNotification(noti); err != nil {
			logger.Error(err)
		}
	}()

	if err := payClient.VerifyNotification(noti); err != nil {
		logger.WithFields(logrus.Fields{
			"event":   "VerifyNotification",
			"orderId": noti.FTCOrderID,
		}).Error(err)

		if _, err := w.Write([]byte(resp.OK())); err != nil {
			logger.Error(err)
		}
		return
	}

	payResult, err := subscription.NewPaymentResultWx(noti)
	if err != nil {
		logger.Error(err)
		if _, err := w.Write([]byte(resp.OK())); err != nil {
			logger.Error()
		}

		return
	}

	confirmedOrder, result := router.subEnv.ConfirmOrder(payResult)

	if result != nil {
		go func() {
			_ = router.subEnv.SaveConfirmationResult(
				result.Schema(payResult.OrderID))
		}()

		if result.Retry {
			if _, err := w.Write([]byte(resp.NotOK(result.Error()))); err != nil {
				logger.Error(err)
			}
		} else {
			if _, err := w.Write([]byte(resp.OK())); err != nil {
				logger.Error(err)
			}
		}

		return
	}

	go func() {
		if err := router.sendConfirmationEmail(confirmedOrder); err != nil {
			logger.Error(err)
		}
	}()

	if _, err := w.Write([]byte(resp.OK())); err != nil {
		logger.Error(err)
	}
}

// OrderQuery implements 查询订单
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
// {orderId}?app_id=<string>
func (router WxPayRouter) OrderQuery(w http.ResponseWriter, req *http.Request) {
	logger := logrus.WithFields(logrus.Fields{
		"trace": "WxPayRouter.OrderQuery()",
	})

	// Get ftc order id from URL
	orderID, err := GetURLParam(req, "orderId").ToString()

	if err != nil {
		logger.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	appID := getWxAppID(req)

	payClient, err := router.clients.GetClientByAppID(appID)
	if err != nil {
		logger.Error(err)
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
		logger.Error(err)
		_ = render.New(w).InternalServerError(err.Error())
		return
	}

	logger.WithField("param", respParams).Infof("Wechat order found")

	// Response:
	// {message: "", {field: status, code: fail} }
	// {message: "", {field: result, code: "ORDERNOTEXIST" | "SYSTEMERROR"} }
	resp := wechat.NewOrderQueryResp(respParams)
	go func() {
		if err := router.subEnv.SaveWxQueryResp(resp); err != nil {
			logger.Error(err)
		}
	}()

	if r := resp.Validate(payClient.GetApp()); r != nil {
		logger.Info("Response invalid")

		if r.Field == "result" && r.Code == "ORDERNOTEXIST" {
			_ = render.New(w).NotFound()
			return
		}

		_ = render.New(w).Unprocessable(r)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, resp.ToQueryResult())
}
