package controller

import (
	"database/sql"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	cache "github.com/patrickmn/go-cache"
	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
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
	model  model.Env
}

// NewAliRouter create a new instance of AliPayRouter
func NewAliRouter(db *sql.DB, c *cache.Cache) AliPayRouter {
	appID := os.Getenv("ALIPAY_APP_ID")

	host := os.Getenv("HANQI_SMTP_HOST")
	user := os.Getenv("HANQI_SMTP_USER")
	portStr := os.Getenv("HANQI_SMTP_PORT")
	pass := os.Getenv("HANQI_SMTP_PASS")

	port, _ := strconv.Atoi(portStr)

	// Ali's public key is used to verify alipay's response.
	publicKey, err := ioutil.ReadFile("alipay_public_key.pem")
	if err != nil {
		logger.WithField("trace", "NewAliRouter").Error(err)
		os.Exit(1)
	}

	// Private key is used to sign our data that will be sent to alipay.
	privateKey, err := ioutil.ReadFile("ftc_private_key.pem")
	if err != nil {
		logger.WithField("location", "NewAliRouter").Error(err)
		os.Exit(1)
	}

	client := alipay.New(appID, string(publicKey), string(privateKey), true)

	return AliPayRouter{
		appID:  appID,
		client: client,
		model: model.Env{
			DB:      db,
			Cache:   c,
			Postman: postoffice.NewPostman(host, port, user, pass),
		},
	}
}

// AppOrder creates an alipay order for native app.
func (router AliPayRouter) AppOrder(w http.ResponseWriter, req *http.Request) {
	// Get member tier and billing cycle from url
	tierKey := getURLParam(req, "tier").toString()
	cycleKey := getURLParam(req, "cycle").toString()

	if tierKey == "" || cycleKey == "" {
		view.Render(w, view.NewBadRequest(msgInvalidURI))
		return
	}

	plan, err := router.model.GetCurrentPricing().FindPlan(tierKey, cycleKey)

	if err != nil {
		logger.WithField("trace", "AliAppOrder").Error(err)

		view.Render(w, view.NewBadRequest(msgInvalidURI))
		return
	}

	logger.WithField("trace", "AliAppOrder").Infof("Subscritpion plan: %+v", plan)

	// Get user id from request header
	userID := req.Header.Get(userIDKey)
	unionID := req.Header.Get(unionIDKey)
	var loginMethod enum.LoginMethod
	if userID != "" {
		loginMethod = enum.LoginMethodEmail
	} else if unionID != "" {
		loginMethod = enum.LoginMethodWx
		userID = unionID
	}

	subs := paywall.NewAlipaySubs(userID, plan, loginMethod)

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

	param := subs.AliAppPayParam(plan.Description)

	// Call URLValues to generate alipay required data structure and sign it.
	values, err := router.client.URLValues(param)

	logger.WithField("trace", "AliAppOrder").Infof("App pay param: %+v\n", values)

	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	resp := subs.AliAppPayResp(values.Encode())

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
		logger.WithField("location", "AliNotification").Infof("Order %s paid", noti.OutTradeNo)

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
			logger.WithField("trace", "AliNotification").Info("Suscription order is not found")
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

	go router.model.SendConfirmationLetter(confirmedSubs)

	w.Write([]byte(success))
}
