package controller

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/models/wechat"
	"gitlab.com/ftchinese/subscription-api/repository"
	"net/http"
)

// WxPayRouter wraps wxpay and alipay sdk instances.
type WxPayRouter struct {
	clients map[string]wechat.Client
	PayRouter
}

// NewWxRouter creates a new instance or OrderRouter
func NewWxRouter(m repository.Env, p postoffice.Postman) WxPayRouter {
	r := WxPayRouter{
		clients: createWxpayClients(),
	}
	r.env = m
	r.postman = p

	return r
}

// Select a pay client based on trade type.
func (router WxPayRouter) selectClient(tradeType wechat.TradeType) (wechat.Client, error) {
	var appID string

	switch tradeType {
	// Desktop and mobile browser
	case wechat.TradeTypeDesktop,
		wechat.TradeTypeMobile:
		appID = wxAppWebPay

	// Wechat in-house browser
	case wechat.TradeTypeJSAPI:
		appID = wxAppWeBrowserPay

	// Native app.
	case wechat.TradeTypeApp:
		appID = wxAppNativeApp
	}

	if appID == "" {
		return wechat.Client{}, errors.New("wechat app id is empty")
	}

	c, ok := router.clients[appID]

	if !ok {
		return c, fmt.Errorf("wxpay client for %s not found", appID)
	}

	return c, nil
}

// Select a pay client by app id.
func (router WxPayRouter) findClient(appID string) (wechat.Client, error) {
	c, ok := router.clients[appID]
	if !ok {
		return c, fmt.Errorf("wxpay client for %s not found", appID)
	}

	return c, nil
}

// PlaceOrder creates order for wechat pay.
func (router WxPayRouter) PlaceOrder(tradeType wechat.TradeType) http.HandlerFunc {
	logger := logrus.WithFields(logrus.Fields{
		"trace": "WxPayRouter.PlaceOrder",
		"type":  tradeType.String(),
	})

	return func(w http.ResponseWriter, req *http.Request) {
		logger.Info("Start placing a wechat order")

		clientApp := util.NewClientApp(req)

		logger.Infof("Client app: %+v", clientApp)

		if err := models.AllowAndroidPurchase(clientApp); err != nil {
			view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		// Find the client to user for wxpay
		//var appID string
		// openID is required for JSAPI pay.
		//var openID string
		openID, _ := util.GetJSONString(req.Body, "openId")

		if tradeType == wechat.TradeTypeJSAPI && openID == "" {
			logger.Error("Requesting JSAPI without providing open id")
			r := view.NewReason()
			r.Field = "openId"
			r.Code = view.CodeMissingField
			r.SetMessage("You must provide open id to use wechat js api")
			view.Render(w, view.NewUnprocessable(r))
			return
		}

		payClient, err := router.selectClient(tradeType)

		if err != nil {
			logger.Error(err)
			view.Render(w, view.NewInternalError(err.Error()))
			return
		}

		// Get ftc user id or wechat union id.
		user, _ := GetUserID(req.Header)

		// Try to find a plan based on the tier and cycle.
		plan, err := router.findPlan(req)
		// If pricing plan is not found.
		if err != nil {
			logger.Error(err)
			view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		// Save this subscription order.
		order, err := router.createOrder(
			user,
			plan,
			enum.PayMethodWx,
			clientApp,
			null.StringFrom(payClient.GetApp().AppID),
		)

		if err != nil {
			logger.Error(err)
			router.handleOrderErr(w, err)
			return
		}

		logger.Infof("Created order: %+v", order)

		// Wxpay specific handling.
		// Prepare the data used to obtain prepay order from wechat.
		unifiedOrder := wechat.UnifiedOrder{
			Body:        plan.Title,
			OrderID:     order.ID,
			Price:       order.AmountInCent(),
			IP:          clientApp.UserIP.String,
			CallbackURL: router.wxCallbackURL(),
			TradeType:   tradeType,
			ProductID:   plan.NamedKey(),
			OpenID:      openID,
		}
		// Build Wechat pay parameters.
		// openID will be added conditionally.
		param := unifiedOrder.ToParam()

		logger.WithField("param", param).Info("Create parameter for wechat")

		// Send order to wx
		// UnifiedOrder checks if `return_code` is SUCCESS/FAIL,
		// validate the signature
		// You have to check if return_code == SUCCESS, appid, mch_id, result_code are valid.
		resp, err := payClient.UnifiedOrder(param)

		if err != nil {
			logger.Error(err)

			view.Render(w, view.NewBadRequest(err.Error()))

			return
		}

		// Convert wxpay's map to struct for easy manipulation.
		uor := wechat.NewUnifiedOrderResp(order.ID, resp)

		go func() {
			if err := router.env.SavePrepayResp(uor); err != nil {
				logger.Error(err)
			}
		}()

		if r := uor.Validate(payClient.GetApp()); r != nil {
			logger.Info("Invalid unified order response")
			view.Render(w, view.NewUnprocessable(r))
			return
		}

		switch tradeType {
		// Desktop returns a url that can be turned to QR code
		case wechat.TradeTypeDesktop:
			order := wechat.BuildDesktopOrder(uor, order)
			view.Render(w, view.NewResponse().SetBody(order))

		// Mobile returns a url which is redirect in browser
		case wechat.TradeTypeMobile:
			order := wechat.BuildMobileOrder(uor, order)
			view.Render(w, view.NewResponse().SetBody(order))

		// Create the json data used by js api
		case wechat.TradeTypeJSAPI:
			//browserPay := uor.ToWxBrowserPay(subs)
			//sign := payClient.Sign(browserPay.Params())
			order := payClient.BuildInAppBrowserOrder(uor, order)
			view.Render(w, view.NewResponse().SetBody(order))

		// Create the json data used by native app.
		case wechat.TradeTypeApp:
			//appPay := uor.ToAppPay(subs)
			//sign := payClient.Sign(appPay.Params())
			order := payClient.BuildAppOrder(uor, order)
			view.Render(w, view.NewResponse().SetBody(order))
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
	payClient, err := router.findClient(noti.AppID.String)

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
		if err := router.env.SaveWxNotification(noti); err != nil {
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

	payResult, err := noti.GetPaymentResult()
	if err != nil {
		logger.Error(err)
		if _, err := w.Write([]byte(resp.OK())); err != nil {
			logger.Error()
		}

		return
	}

	confirmedOrder, result := router.confirmPayment(payResult)

	if result != nil {
		logger.Error(err)

		go func() {
			if err := router.env.SaveConfirmationResult(result); err != nil {
				logger.Error(err)
			}
		}()

		if result.Retry {
			if _, err := w.Write([]byte(resp.NotOK(err.Error()))); err != nil {
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

	go func() {
		if err := router.env.SaveConfirmationResult(paywall.NewConfirmationSucceeded(noti.FTCOrderID.String)); err != nil {
			logger.Error(err)
		}
	}()

	if _, err := w.Write([]byte(resp.OK())); err != nil {
		logger.Error(err)
	}
}

// OrderQuery implements 查询订单
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
// Only transaction_id or out_trade_no is required.
func (router WxPayRouter) OrderQuery(w http.ResponseWriter, req *http.Request) {
	logger := logrus.WithFields(logrus.Fields{
		"trace": "WxPayRouter.OrderQuery()",
	})

	orderID, err := GetURLParam(req, "orderId").ToString()

	if err != nil {
		logger.Error(err)
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// Find out which app is used to create this order.
	appID := req.Header.Get(appIDKey)
	// For backward compatibility with Android <= 2.0.4
	if appID == "" {
		appID = wxAppNativeApp
	}

	payClient, err := router.findClient(appID)
	if err != nil {
		logger.Error(err)
		view.Render(w, view.NewBadRequest(err.Error()))
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

		view.Render(w, view.NewInternalError(err.Error()))

		return
	}

	logger.WithField("param", respParams).Infof("Wechat order found")

	// Response:
	// {message: "", {field: status, code: fail} }
	// {message: "", {field: result, code: "ORDERNOTEXIST" | "SYSTEMERROR"} }
	resp := wechat.NewOrderQueryResp(respParams)
	go func() {
		if err := router.env.SaveWxQueryResp(resp); err != nil {
			logger.Error(err)
		}
	}()

	if r := resp.Validate(payClient.GetApp()); r != nil {
		logger.Info("Response invalid")

		if r.Field == "result" && r.Code == "ORDERNOTEXIST" {
			view.Render(w, view.NewNotFound())
			return
		}

		view.Render(w, view.NewUnprocessable(r))
		return
	}

	view.Render(w, view.NewResponse().SetBody(resp.ToQueryResult()))
}
