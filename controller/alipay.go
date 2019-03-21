package controller

import (
	"database/sql"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/guregu/null"
	"github.com/smartwalle/alipay"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/ali"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
	"net/http"
	"os"
)

const (
	success       = "success"
	fail          = "fail"
	tradeFinished = "TRADE_FINISHED"
	tradeSuccess  = "TRADE_SUCCESS"
	tradePending  = "WAIT_BUYER_PAY"
	tradeClosed   = "TRADE_CLOSED"
)

// AliPayRouter handles alipay request
type AliPayRouter struct {
	appID  string
	client *alipay.AliPay
	PayRouter
}

// NewAliRouter create a new instance of AliPayRouter
func NewAliRouter(m model.Env, p postoffice.Postman, sandbox bool) AliPayRouter {
	var app ali.App

	if err := viper.UnmarshalKey("alipay", &app); err != nil {
		logger.WithField("trace", "NewAliRouter").Error(err)
		os.Exit(1)
	}

	if err := app.Ensure(); err != nil {
		logger.WithField("trace", "NewAliRouter").Error(err)
		os.Exit(1)
	}

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

// AppOrder creates an alipay order for native app.
func (router AliPayRouter) AppOrder(w http.ResponseWriter, req *http.Request) {
	// Get member tier and billing cycle from url
	//tierKey := getURLParam(req, "tier").toString()
	//cycleKey := getURLParam(req, "cycle").toString()
	//
	//if tierKey == "" || cycleKey == "" {
	//	view.Render(w, view.NewBadRequest(msgInvalidURI))
	//	return
	//}

	tier, err := GetURLParam(req, "tier").ToString()
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	cycle, err := GetURLParam(req, "cycle").ToString()
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	plan, err := router.model.GetCurrentPricing().FindPlan(tier, cycle)

	if err != nil {
		logger.WithField("trace", "AliAppOrder").Error(err)

		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	logger.WithField("trace", "AliAppOrder").Infof("Subscription plan: %+v", plan)

	// Get user id from request header
	uID := req.Header.Get(userIDKey)
	wID := req.Header.Get(unionIDKey)

	userID := null.NewString(uID, uID != "")
	unionID := null.NewString(wID, wID != "")

	logger.WithField("trace", "AliAppOrder").Infof("FTC id: %+v, wechat id: %+v", userID, unionID)

	subs, err := paywall.NewAlipaySubs(userID, unionID, plan)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	logger.WithField("trace", "AliAppOrder").Infof("User created order: %+v", subs)

	ok, err := router.model.IsSubsAllowed(subs)
	// err = ar.model.PlaceOrder(subs, app)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}
	if !ok {
		view.Render(w, view.NewForbidden("Already a subscribed user and not within allowed renewal period."))
		return
	}

	// Save the subscription
	app := gorest.NewClientApp(req)
	err = router.model.SaveSubscription(subs, app)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	param := router.aliAppPayParam(plan.Description, subs)

	// Call URLValues to generate alipay required data structure and sign it.
	values, err := router.client.URLValues(param)

	logger.WithField("trace", "AliAppOrder").Infof("App pay param: %+v\n", values)

	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	resp := ali.NewAppPayResp(subs, values.Encode())

	view.Render(w, view.NewResponse().SetBody(resp))
}

// Notification handles alipay server-side notification.
func (router AliPayRouter) Notification(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()

	if err != nil {
		logger.WithField("trace", "AliNotification").Error(err)

		w.Write([]byte(fail))
		return
	}

	// If err is nil, then the signature is verified.
	noti, err := router.client.GetTradeNotification(req)

	logger.WithField("trace", "AliNotification").Infof("%+v", noti)

	if err != nil {
		logger.WithField("trace", "AliNotification").Error(err)

		w.Write([]byte(fail))
		return
	}

	logger.WithField("trace", "AliNotification").Infof("Ali notification data: %+v", noti)

	// 4、验证app_id是否为该商户本身
	if noti.AppId != router.appID {
		logger.WithField("trace", "AliNotification").Info("AppID does not match")

		w.Write([]byte(fail))
		return
	}

	go router.model.SaveAliNotification(*noti)

	// 在支付宝的业务通知中，只有交易通知状态为TRADE_SUCCESS或TRADE_FINISHED时，支付宝才会认定为买家付款成功。
	switch noti.TradeStatus {
	case tradeSuccess, tradeFinished:
		logger.WithField("trace", "AliNotification").Infof("Order %s paid", noti.OutTradeNo)

	case tradePending:
		logger.WithField("trace", "AliNotification").Info("Payment pending")
		w.Write([]byte(fail))
		return

	case tradeClosed:
		logger.WithField("trace", "AliNotification").Info("Transaction closed. Money not earned. :-<")
		w.Write([]byte(success))
		return

	default:
		logger.WithField("trace", "AliNotification").Info("Unknown trade status")
		w.Write([]byte(success))
		return
	}

	orderID := noti.OutTradeNo
	// 1、商户需要验证该通知数据中的out_trade_no是否为商户系统中创建的订单号
	subs, err := router.model.FindSubscription(orderID)

	// If the order does not exist, tell ali success;
	// If err is not `not found`, tell ali to resend.
	if err != nil {
		if err == sql.ErrNoRows {
			logger.WithField("trace", "AliNotification").Info("Subscription order is not found")
			w.Write([]byte(success))
		}
		w.Write([]byte(fail))
		return
	}

	// 2、判断total_amount是否确实为该订单的实际金额（即商户订单创建时的金额）
	if subs.AliNetPrice() != noti.TotalAmount {
		logger.WithField("trace", "AliNotification").Infof("Expected net price: %s, actually received: %s", subs.AliNetPrice(), noti.TotalAmount)

		w.Write([]byte(success))

		return
	}

	// If this order already confirmed.
	if !subs.ConfirmedAt.IsZero() {
		w.Write([]byte(success))
		return
	}

	confirmedSubs, err := router.model.ConfirmPayment(orderID, util.ParseAliTime(noti.GmtPayment))

	if err != nil {
		switch err {
		case model.ErrOrderNotFound, model.ErrAlreadyConfirmed:
			w.Write([]byte(success))
			return
		default:
			w.Write([]byte(fail))
			return
		}
	}

	go router.sendConfirmationEmail(confirmedSubs)

	w.Write([]byte(success))
}
