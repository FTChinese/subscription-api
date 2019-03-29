package controller

import (
	"database/sql"
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
	"net/url"
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

func (router AliPayRouter) PlaceOrder(w http.ResponseWriter, req *http.Request) {
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

	// Save the subscription
	app := util.NewClientApp(req)
	err = router.model.SaveSubscription(subs, app)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	param := router.aliWebPayParam(plan.Description, subs)

	payURL, err := router.client.TradePagePay(param)

	logger.WithField("trace", "AliAppOrder").Infof("App pay param: %+v\n", payURL)

	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	resp := ali.NewDesktopWebPay(subs, payURL)

	view.Render(w, view.NewResponse().SetBody(resp))
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
	app := util.NewClientApp(req)
	err = router.model.SaveSubscription(subs, app)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	param := router.aliAppPayParam(plan.Description, subs)

	// Call URLValues to generate alipay required data structure and sign it.
	//values, err := router.client.URLValues(param)

	payURL, err := router.client.TradeAppPay(param)

	logger.WithField("trace", "AliAppOrder").Infof("App pay param: %+v\n", payURL)

	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	resp := ali.NewAppPayResp(subs, payURL)

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

// RedirectNextUser redirect alipay result to next-user app.
// {
//  "app_id": [
//    "2018053060263354"
//  ],
//  "auth_app_id": [
//    "2018053060263354"
//  ],
//  "charset": [
//    "utf-8"
//  ],
//  "method": [
//    "alipay.trade.page.pay.return"
//  ],
//  "out_trade_no": [
//    "FT99a55d609736c4fa"
//  ],
//  "seller_id": [
//    "2088521304936335"
//  ],
//  "sign": [
//    "CFOT1unSWR5tenWbfOndNQbgfgeql5QyJqdZzVFCMIw2KERb1jjV3lqqPPB9stc8A8yvyLZnyncSqzyoD0Ss3bfvnUuxSEpNHyANXgNn+uBTa4Niqc/bRBlH4k2RhE2SjkYDRpQ1rFEcM0/92rAdfZ35QSPquirrQNut5C7Tnt97A2nuw+YKuUubZ+tixKLWOMy9ADJAW+1ZDJNjX4bJs2RpIEyVHD8XViQ1M2e8LsV4C+lG011cGTGQe4QmSA/AeS5FPSeMaAKVvvfP/S9N+Ddhctv54+orxBisQ1YNM+Wn22IpJ4Ra5G6FVm+AWLjTC0yAeEugKm9ed7Mgw2XMRQ=="
//  ],
//  "sign_type": [
//    "RSA2"
//  ],
//  "timestamp": [
//    "2019-03-29 14:06:58"
//  ],
//  "total_amount": [
//    "0.01"
//  ],
//  "trade_no": [
//    "2019032922001440031023691208"
//  ],
//  "version": [
//    "1.0"
//  ]
//}
func (router AliPayRouter) RedirectNextUser(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	logger.WithField("trace", "RedirectNextUser").Infof("Query parameters: %s", req.Form)

	baseUrl := "http://next.ftchinese.com/user/subscription/alipay/callback?"

	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	ok, err := router.client.VerifySign(req.Form)
	if err != nil {
		query := url.Values{}
		query.Set("error", "invalid_signature")
		query.Set("error_description", err.Error())

		http.Redirect(
			w,
			req,
			baseUrl+query.Encode(),
			http.StatusFound)
		return
	}

	if !ok {
		query := url.Values{}
		query.Set("error", "invalid_signature")

		http.Redirect(
			w,
			req,
			baseUrl+query.Encode(),
			http.StatusFound)
		return
	}

	http.Redirect(
		w,
		req,
		baseUrl+req.Form.Encode(),
		http.StatusFound)
}
