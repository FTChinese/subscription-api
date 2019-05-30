package controller

import (
	"database/sql"
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"github.com/pkg/errors"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/wechat"
	"net/http"
	"time"
)

// WxPayRouter wraps wxpay and alipay sdk instances.
type WxPayRouter struct {
	clients map[string]wechat.Client
	PayRouter
}

// NewWxRouter creates a new instance or OrderRouter
func NewWxRouter(m model.Env, p postoffice.Postman, sandbox bool) WxPayRouter {
	r := WxPayRouter{
		clients: createWxpayClients(),
	}
	r.sandbox = sandbox
	r.model = m
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
		appID = wxAppMobileFTC

	// Wechat in-house browser
	case wechat.TradeTypeJSAPI:
		appID = wxAppFTCSupport

	// Native app.
	case wechat.TradeTypeApp:
		appID = wxAppMobileSubs
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
	return func(w http.ResponseWriter, req *http.Request) {
		// Find the client to user for wxpay
		//var appID string
		// openID is required for JSAPI pay.
		//var openID string
		openID, _ := util.GetJSONString(req.Body, "openId")

		if tradeType == wechat.TradeTypeJSAPI && openID == "" {
			r := view.NewReason()
			r.Field = "openId"
			r.Code = view.CodeMissingField
			r.SetMessage("You must provide open id to use wechat js api")
			view.Render(w, view.NewUnprocessable(r))
			return
		}

		payClient, err := router.selectClient(tradeType)

		if err != nil {
			view.Render(w, view.NewInternalError(err.Error()))
			return
		}

		// Get ftc user id or wechat union id.
		user, _ := GetUser(req.Header)

		// Try to find a plan based on the tier and cycle.
		plan, err := router.findPlan(req)
		// If pricing plan is not found.
		if err != nil {
			logger.WithField("trace", "UnifiedOrder").Error(err)
			view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		// Create an subscription order for this user
		// base on chosen plan.
		//subs, err := router.createOrder(user, plan)
		//if err != nil {
		//	router.handleOrderErr(w, err)
		//	return
		//}
		//subs = subs.WithWxpay(payClient.GetApp().AppID)

		// Save this subscription order.
		clientApp := util.NewClientApp(req)
		//err = router.model.SaveSubscription(subs, clientApp)
		//if err != nil {
		//	view.Render(w, view.NewDBFailure(err))
		//	return
		//}

		subs, err := router.model.CreateOrder(
			user,
			plan,
			enum.PayMethodWx,
			clientApp,
			null.StringFrom(payClient.GetApp().AppID),
		)
		if err != nil {
			router.handleOrderErr(w, err)
			return
		}

		// Wxpay specific handling.
		// Prepare the data used to obtain prepay order from wechat.
		unifiedOrder := wechat.UnifiedOrder{
			Body:        plan.Description,
			OrderID:     subs.OrderID,
			Price:       subs.WxNetPrice(),
			IP:          clientApp.UserIP.String,
			CallbackURL: router.wxCallbackURL(),
			TradeType:   tradeType,
			ProductID:   plan.ProductID(),
			OpenID:      openID,
		}
		// Build Wechat pay parameters.
		// openID will be added conditionally.
		param := unifiedOrder.ToParam()

		logger.WithField("trace", "UnifiedOrder").Infof("Unified order params: %+v", param)

		// Send order to wx
		// UnifiedOrder checks if `return_code` is SUCCESS/FAIL,
		// validate the signature
		// You have to check if return_code == SUCCESS, appid, mch_id, result_code are valid.
		resp, err := payClient.UnifiedOrder(param)

		if err != nil {
			logger.WithField("trace", "UnifiedOrder").Error(err)

			view.Render(w, view.NewBadRequest(err.Error()))

			return
		}

		// Convert wxpay's map to struct for easy manipulation.
		uor := wechat.NewUnifiedOrderResp(resp)

		go router.model.SavePrepayResp(subs.OrderID, uor)

		if r := uor.Validate(payClient.GetApp()); r != nil {
			view.Render(w, view.NewUnprocessable(r))
			return
		}

		switch tradeType {
		// Desktop returns a url that can be turned to QR code
		case wechat.TradeTypeDesktop:
			desktopPay := uor.ToDesktopPay(subs)
			view.Render(w, view.NewResponse().SetBody(desktopPay))

		// Mobile returns a url which is redirect in browser
		case wechat.TradeTypeMobile:
			mobilePay := uor.ToMobilePay(subs)
			view.Render(w, view.NewResponse().SetBody(mobilePay))

		// Create the json data used by js api
		case wechat.TradeTypeJSAPI:
			browserPay := uor.ToWxBrowserPay(subs)
			sign := payClient.Sign(browserPay.Params())
			view.Render(w, view.NewResponse().SetBody(browserPay.WithHash(sign)))

		// Create the json data used by native app.
		case wechat.TradeTypeApp:
			appPay := uor.ToAppPay(subs)
			sign := payClient.Sign(appPay.Params())
			view.Render(w, view.NewResponse().SetBody(appPay.WithHash(sign)))
		}
	}
}

// UnifiedOrder implements 统一下单.
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_1
//func (router WxPayRouter) AppOrder(w http.ResponseWriter, req *http.Request) {
//
//	payClient, err := router.clients.Find(wxAppMobileSubs)
//	if err != nil {
//		view.Render(w, view.NewInternalError(err.Error()))
//		return
//	}
//
//	// Try to find a plan based on the tier and cycle.
//	plan, err := router.findPlan(req)
//
//	// If pricing plan is not found.
//	if err != nil {
//		logger.WithField("trace", "UnifiedOrder").Error(err)
//		view.Render(w, view.NewBadRequest(err.Error()))
//		return
//	}
//
//	logger.WithField("trace", "UnifiedOrder").Infof("Subscription plan: %+v", plan)
//
//	// Get user id from request header.
//	// If user id is found, it means user is subscribing with FTC account;
//	// if union id is found, it means user is subscribing with Wechat account;
//	user, err := GetUser(req.Header)
//
//	subs, err := paywall.NewWxpaySubs(user, plan)
//	if err != nil {
//		view.Render(w, view.NewBadRequest(err.Error()))
//		return
//	}
//	subs.WxAppID = null.StringFrom(wxAppMobileSubs)
//
//	ok, err := router.model.IsSubsAllowed(subs)
//	if err != nil {
//		view.Render(w, view.NewDBFailure(err))
//		return
//	}
//	if !ok {
//		view.Render(w, view.NewForbidden("Already a subscribed user and not within allowed renewal period."))
//		return
//	}
//
//	// Save this subscription order.
//	clientApp := util.NewClientApp(req)
//	err = router.model.SaveSubscription(subs, clientApp)
//	if err != nil {
//		view.Render(w, view.NewDBFailure(err))
//		return
//	}
//
//	unifiedOrder := wechat.UnifiedOrder{
//		Body:        plan.Description,
//		OrderID:     subs.OrderID,
//		Price:       subs.WxNetPrice(),
//		IP:          clientApp.UserIP.String,
//		CallbackURL: router.wxCallbackURL(),
//		TradeType:   wechat.TradeTypeApp,
//		ProductID:   plan.ProductID(),
//	}
//	// Build Wechat pay parameters.
//	//param := router.wxUniOrderParam(plan.Description, clientApp.UserIP.String, subs)
//	param := unifiedOrder.ToParam()
//
//	logger.WithField("trace", "UnifiedOrder").Infof("Unifed order params: %+v", param)
//
//	// Send order to wx
//	resp, err := payClient.UnifiedOrder(param)
//
//	if err != nil {
//		logger.WithField("trace", "UnifiedOrder").Error(err)
//
//		view.Render(w, view.NewBadRequest(err.Error()))
//
//		return
//	}
//
//	uor := wechat.NewUnifiedOrderResp(resp)
//
//	// Log the response.
//	go router.model.SavePrepayResp(subs.OrderID, uor)
//
//	//if r := router.mSubsClient.ValidateResponse(resp); r != nil {
//	//	view.Render(w, view.NewUnprocessable(r))
//	//	return
//	//}
//
//	if r := uor.Validate(payClient.GetApp()); r != nil {
//		view.Render(w, view.NewUnprocessable(r))
//		return
//	}
//
//	//prepay := router.mSubsClient.NewPrepay(resp.GetString("prepay_id"), subs)
//	appPay := uor.ToLegacyAppPay(subs)
//
//	sign := payClient.Sign(appPay.Param())
//
//	view.Render(w, view.NewResponse().SetBody(appPay.WithHash(sign)))
//}

// Notification implements 支付结果通知
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_7&index=3
func (router WxPayRouter) Notification(w http.ResponseWriter, req *http.Request) {

	resp := wxpay.Notifies{}

	params, err := wechat.DecodeXML(req.Body)
	if err != nil {
		logger.WithField("trace", "Notification").Error(err)

		w.Write([]byte(resp.NotOK(err.Error())))

		return
	}

	logger.WithField("trace", "WxpayNotification").Infof("%+v", params)

	noti := wechat.NewNotification(params)

	err = noti.IsStatusValid()
	if err != nil {
		logger.WithField("trace", "WxpayNotification").Error(err)
		w.Write([]byte(resp.OK()))
		return
	}

	// Try to find out which app is in charge of the response.
	payClient, err := router.findClient(noti.AppID.String)

	//params, err := router.mSubsClient.ParseResponse(req.Body)
	if err != nil {
		logger.WithField("trace", "Notification").Error(err)

		w.Write([]byte(resp.NotOK(err.Error())))

		return
	}

	// Log the response, regardless of whether it is an error
	// or not.
	go router.model.SaveWxNotification(noti)

	if err := payClient.VerifyNotification(noti); err != nil {
		logger.WithField("trace", "Notification").Error(err)
		w.Write([]byte(resp.OK()))
		return
	}

	// Verify appid, mch_id, trade_type, total_fee.
	//if r := router.mSubsClient.ValidateResponse(params); r != nil {
	//	logger.WithField("trace", "Notification").Error(r.GetMessage())
	//	w.Write([]byte(resp.OK()))
	//
	//	return
	//}

	// Get out_trade_no to retrieve order.
	// Check the order's confirmed_utc field.
	// If confirmed_utc is empty, get time_end from params and set confirmed_utc to it.
	//orderID := params.GetString("out_trade_no")
	charge, err := router.model.FindSubsCharge(noti.FTCOrderID.String)
	if err != nil {
		if err == sql.ErrNoRows {
			w.Write([]byte(resp.OK()))
			return
		}
		w.Write([]byte(resp.NotOK(err.Error())))
		return
	}

	if !noti.IsPriceMatched(charge.WxNetPrice()) {
		logger.WithField("trace", "WxpayNotification").Errorf("Wxpay total amount does not match. Expected: %d, actual: ", charge.WxNetPrice(), noti.TotalFee.Int64)

		w.Write([]byte(resp.OK()))
		return
	}

	if charge.IsConfirmed {
		w.Write([]byte(resp.OK()))
		return
	}
	// updatedSubs
	//timeEnd := params.GetString("time_end")

	confirmedAt, err := util.ParseWxTime(noti.TimeEnd.String)
	if err != nil {
		confirmedAt = time.Now()
	}
	confirmedSubs, err := router.model.ConfirmPayment(noti.FTCOrderID.String, confirmedAt)

	if err != nil {
		logger.WithField("trace", "Notification").Error(err)

		switch err {
		case model.ErrDenyRetry:
			w.Write([]byte(resp.OK()))
			return

		case model.ErrAllowRetry:
			w.Write([]byte(resp.NotOK(err.Error())))
			return

		default:
			w.Write([]byte(resp.NotOK(err.Error())))
			return
		}
	}

	// Send a letter to this user.
	go router.sendConfirmationEmail(confirmedSubs)

	w.Write([]byte(resp.OK()))
}

// OrderQuery implements 查询订单
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
// Only transaction_id or out_trade_no is required.
func (router WxPayRouter) OrderQuery(w http.ResponseWriter, req *http.Request) {
	orderID, err := GetURLParam(req, "orderId").ToString()

	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// Find out which app is used to create this order.
	appID := req.Header.Get(appIDKey)
	// For backward compatibility with Android <= 2.0.4
	if appID == "" {
		appID = wxAppMobileSubs
	}

	payClient, err := router.findClient(appID)
	if err != nil {
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
		logger.WithField("trace", "OrderQuery").Error(err)

		view.Render(w, view.NewInternalError(err.Error()))

		return
	}

	logger.WithField("trace", "OrderQuery").Infof("Order query result: %+v", respParams)

	// Response:
	// {message: "", {field: status, code: fail} }
	// {message: "", {field: result, code: "ORDERNOTEXIST" | "SYSTEMERROR"} }
	resp := wechat.NewOrderQueryResp(respParams)
	go router.model.SaveWxQueryResp(resp)

	if r := resp.Validate(payClient.GetApp()); r != nil {
		if r.Field == "result" && r.Code == "ORDERNOTEXIST" {
			view.Render(w, view.NewNotFound())
			return
		}

		view.Render(w, view.NewUnprocessable(r))
		return
	}

	view.Render(w, view.NewResponse().SetBody(resp.ToQueryResult()))
}
