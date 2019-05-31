package controller

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/guregu/null"
	"github.com/sirupsen/logrus"
	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/ali"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/util"
	"net/http"
	"net/url"
)

const (
	success = "success"
	fail    = "fail"
)

// AliPayRouter handles alipay request
type AliPayRouter struct {
	appID  string
	client *alipay.AliPay
	PayRouter
}

// NewAliRouter create a new instance of AliPayRouter
func NewAliRouter(m model.Env, p postoffice.Postman, sandbox bool) AliPayRouter {

	app := getAliPayApp()

	client := alipay.New(app.ID, app.PublicKey, app.PrivateKey, true)

	r := AliPayRouter{
		appID:  app.ID,
		client: client,
	}
	r.sandbox = sandbox
	r.model = m
	r.postman = p

	return r
}

// PlaceOrder creates an http handler function depending
// on the device platform.
//
// 	POST /<desktop|mobile|app>/{tier}/{cycle}?<return_url=xxx>
// `return_url` parameter is only required for apps running on ftacademy.cn
func (router AliPayRouter) PlaceOrder(kind ali.EntryKind) http.HandlerFunc {
	logger := logrus.WithFields(logrus.Fields{
		"trace": "AliPayRouter.PlaceOrder",
		"type":  kind.String(),
	})

	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			logger.Error(err)
			view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		user, _ := GetUser(req.Header)

		plan, err := router.findPlan(req)

		if err != nil {
			logger.Error(err)

			view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		// Save the subscription
		clientApp := util.NewClientApp(req)

		subs, err := router.model.CreateOrder(
			user,
			plan,
			enum.PayMethodAli,
			clientApp,
			null.String{},
		)
		if err != nil {
			router.handleOrderErr(w, err)
			return
		}

		logger.Infof("Created order: %+v", subs)

		// Alipay specific handling.
		returnURL := req.FormValue("return_url")

		tradePay := alipay.TradePay{
			NotifyURL:   router.aliCallbackURL(),
			ReturnURL:   returnURL,
			Subject:     plan.Description,
			OutTradeNo:  subs.OrderID,
			TotalAmount: subs.AliNetPrice(),
		}

		var payURL = &url.URL{}

		switch kind {
		case ali.EntryApp:
			//param := router.aliAppPayParam(plan.Description, subs)
			param := ali.BuildAppPay(tradePay)
			queryStr, err := router.client.TradeAppPay(param)

			logger.Infof("App pay param: %+v\n", queryStr)

			if err != nil {
				logger.Error(err)
				view.Render(w, view.NewBadRequest(err.Error()))
				return
			}

			resp := ali.NewAppPayResp(subs, queryStr)
			view.Render(w, view.NewResponse().SetBody(resp))
			// For pay from app you should stop here.
			return

		case ali.EntryDesktopWeb:
			//param := router.aliDesktopPayParam(plan.Description, subs)
			param := ali.BuildDesktopPay(tradePay)
			payURL, err = router.client.TradePagePay(param)

		case ali.EntryMobileWeb:
			//param := router.aliWapPayParam(plan.Description, subs)
			param := ali.BuildWapPay(tradePay)
			payURL, err = router.client.TradeWapPay(param)
		}

		// Following handles pay from browser, both desktop and mobile.
		logger.Infof("Ali Web pay param: %+v\n", payURL)

		if err != nil {
			logger.Error(err)
			view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		resp := ali.NewWebPay(subs, payURL)
		view.Render(w, view.NewResponse().SetBody(resp))
	}
}

// AppOrder creates an alipay order for native app.
//func (router AliPayRouter) AppOrder(w http.ResponseWriter, req *http.Request) {
//
//	tier, err := GetURLParam(req, "tier").ToString()
//	if err != nil {
//		view.Render(w, view.NewBadRequest(err.Error()))
//		return
//	}
//
//	cycle, err := GetURLParam(req, "cycle").ToString()
//	if err != nil {
//		view.Render(w, view.NewBadRequest(err.Error()))
//		return
//	}
//
//	plan, err := router.model.GetCurrentPricing().FindPlan(tier, cycle)
//
//	if err != nil {
//		logger.WithField("trace", "AliAppOrder").Error(err)
//
//		view.Render(w, view.NewBadRequest(err.Error()))
//		return
//	}
//
//	logger.WithField("trace", "AliAppOrder").Infof("Subscription plan: %+v", plan)
//
//	// Get user id from request header
//	uID := req.Header.Get(userIDKey)
//	wID := req.Header.Get(unionIDKey)
//
//	userID := null.NewString(uID, uID != "")
//	unionID := null.NewString(wID, wID != "")
//
//	logger.WithField("trace", "AliAppOrder").Infof("FTC id: %+v, wechat id: %+v", userID, unionID)
//
//	subs, err := paywall.NewAlipaySubs(userID, unionID, plan)
//	if err != nil {
//		view.Render(w, view.NewBadRequest(err.Error()))
//		return
//	}
//
//	logger.WithField("trace", "AliAppOrder").Infof("User created order: %+v", subs)
//
//	ok, err := router.model.IsSubsAllowed(subs)
//	// err = ar.model.PlaceOrder(subs, app)
//	if err != nil {
//		view.Render(w, view.NewDBFailure(err))
//		return
//	}
//	if !ok {
//		view.Render(w, view.NewForbidden("Already a subscribed user and not within allowed renewal period."))
//		return
//	}
//
//	// Save the subscription
//	app := util.NewClientApp(req)
//	err = router.model.SaveSubscription(subs, app)
//	if err != nil {
//		view.Render(w, view.NewDBFailure(err))
//		return
//	}
//
//	param := router.aliAppPayParam(plan.Description, subs)
//
//	// Call URLValues to generate alipay required data structure and sign it.
//	//values, err := router.client.URLValues(param)
//
//	queryStr, err := router.client.TradeAppPay(param)
//
//	logger.WithField("trace", "AliAppOrder").Infof("App pay param: %+v\n", queryStr)
//
//	if err != nil {
//		view.Render(w, view.NewBadRequest(err.Error()))
//		return
//	}
//
//	resp := ali.NewAppPayResp(subs, queryStr)
//
//	view.Render(w, view.NewResponse().SetBody(resp))
//}

// Notification handles alipay server-side notification.
func (router AliPayRouter) Notification(w http.ResponseWriter, req *http.Request) {
	logger := logrus.WithFields(logrus.Fields{
		"trace": "AliPayRouter.Notification",
	})

	err := req.ParseForm()
	if err != nil {
		logger.Error(err)

		w.Write([]byte(fail))
		return
	}

	// If err is nil, then the signature is verified.
	noti, err := router.client.GetTradeNotification(req)
	logger.WithFields(logrus.Fields{
		"event": "NotificationBody",
	}).Infof("+%v", noti)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"event": "GetTradeNotification",
		}).Error(err)

		w.Write([]byte(fail))
		return
	}

	// 4、验证app_id是否为该商户本身
	if noti.AppId != router.appID {
		logger.WithFields(logrus.Fields{
			"event":   "AppIDNotMatch",
			"orderId": noti.OutTradeNo,
		}).Infof("Expected %s, actual %s", router.appID, noti.AppId)

		w.Write([]byte(fail))
		return
	}

	go router.model.SaveAliNotification(*noti)

	// 在支付宝的业务通知中，只有交易通知状态为TRADE_SUCCESS或TRADE_FINISHED时，支付宝才会认定为买家付款成功。
	if !ali.IsPaySuccess(noti) {
		logger.WithFields(logrus.Fields{
			"event":   "PaymentFailed",
			"orderId": noti.OutTradeNo,
		}).Infof("Status %s", noti.TradeStatus)

		if ali.ShouldRetry(noti) {
			w.Write([]byte(fail))
			return
		}

		w.Write([]byte(success))
		return
	}

	orderID := noti.OutTradeNo
	// 1、商户需要验证该通知数据中的out_trade_no是否为商户系统中创建的订单号
	charge, err := router.model.FindSubsCharge(orderID)

	// If the order does not exist, tell ali success;
	// If err is not `not found`, tell ali to resend.
	if err != nil {
		logger.WithFields(logrus.Fields{
			"event":   "FindSubsCharge",
			"orderId": noti.OutTradeNo,
		}).Error(err)

		if err == sql.ErrNoRows {
			w.Write([]byte(success))
		}
		w.Write([]byte(fail))
		return
	}

	// 2、判断total_amount是否确实为该订单的实际金额（即商户订单创建时的金额）
	if charge.AliNetPrice() != noti.TotalAmount {
		logger.WithFields(logrus.Fields{
			"event":   "PaymentFailed",
			"orderId": noti.OutTradeNo,
		}).Infof("Expected net price: %s, actually received: %s", charge.AliNetPrice(), noti.TotalAmount)

		w.Write([]byte(success))

		return
	}

	// If this order already confirmed.
	if charge.IsConfirmed {
		logger.WithFields(logrus.Fields{
			"event":   "AlreadyConfirmed",
			"orderId": noti.OutTradeNo,
		}).Info("Duplicate notification since this order is already confirmed.")

		w.Write([]byte(success))
		return
	}

	confirmedSubs, err := router.model.ConfirmPayment(orderID, util.ParseAliTime(noti.GmtPayment))

	if err != nil {
		logger.WithFields(logrus.Fields{
			"event":   "ConfirmOrder",
			"orderId": noti.OutTradeNo,
		}).Error(err)

		switch err {
		case model.ErrDenyRetry:
			w.Write([]byte(success))
			return

		case model.ErrAllowRetry:
			w.Write([]byte(fail))
			return

		default:
			w.Write([]byte(fail))
			return
		}
	}

	logger.WithFields(logrus.Fields{
		"event":   "PaymentFailed",
		"OrderId": noti.OutTradeNo,
	}).Infof("Confirmed at %s, membership from %s to %s", confirmedSubs.ConfirmedAt, confirmedSubs.StartDate, confirmedSubs.EndDate)

	go router.sendConfirmationEmail(confirmedSubs)

	w.Write([]byte(success))
}
