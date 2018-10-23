package controller

import (
	"database/sql"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/model"
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
	isProd bool
	client *alipay.AliPay
	model  model.Env
}

// NewAliRouter create a new instance of AliPayRouter
func NewAliRouter(db *sql.DB, isProd bool) AliPayRouter {
	appID := os.Getenv("ALIPAY_APP_ID")

	publicKey, err := ioutil.ReadFile("alipay_public_key.pem")
	if err != nil {
		logger.WithField("location", "NewAliRouter").Error(err)
		os.Exit(1)
	}

	privateKey, err := ioutil.ReadFile("ftc_private_key.pem")
	if err != nil {
		logger.WithField("location", "NewAliRouter").Error(err)
		os.Exit(1)
	}

	client := alipay.New(appID, string(publicKey), string(privateKey), isProd)

	return AliPayRouter{
		appID:  appID,
		isProd: isProd,
		client: client,
		model:  model.Env{DB: db},
	}
}

// AppOrder an alipay order for native app.
func (ar AliPayRouter) AppOrder(w http.ResponseWriter, req *http.Request) {
	// Get member tier and billing cycle from url
	tierKey := getURLParam(req, "tier").toString()
	cycleKey := getURLParam(req, "cycle").toString()

	tier, err := model.NewTier(tierKey)

	cycle, err := model.NewCycle(cycleKey)

	if err != nil {
		util.Render(w, util.NewBadRequest(msgInvalidURI))
		return
	}

	// Get user id from request header
	userID := req.Header.Get(userIDKey)

	// Find if this user is already subscribed.
	// If a membership is not found, sql.ErrNoRows will be returned.
	// Discard the error.
	member, err := ar.model.FindMember(userID)

	// If membership for this user is found, and is not in the allowed renewal period.
	// Allowed renewal period: current time is within the length of the expiration time minus the requested billing cycle.
	if err == nil && !member.CanRenew(cycle) {
		reason := util.InvalidReason{
			Message: "Already a subscribed user",
			Field:   "order",
			Code:    util.CodeAlreadyExsits,
		}
		util.Render(w, util.NewUnprocessable(reason))
		return
	}

	plan, err := model.NewPlan(tier, cycle)

	if err != nil {
		logger.WithField("location", "AliAppOrder").Error(err)

		util.Render(w, util.NewBadRequest(msgInvalidURI))
		return
	}

	if !ar.isProd {
		plan.Price = 0.01
	}

	logger.WithField("location", "AliAppOrder").Infof("Subscritpion plan: %+v", plan)

	orderID := model.CreateOrderID(plan)

	logger.WithField("location", "AliAppOrder").Infof("Created order: %s", orderID)

	// Get request client metadata
	c := util.NewRequestClient(req)

	// Save this order to db.
	ftcOrder := model.Subscription{
		OrderID:       orderID,
		TierToBuy:     plan.Tier,
		BillingCycle:  plan.Cycle,
		Price:         plan.Price,
		TotalAmount:   plan.Price,
		PaymentMethod: model.Alipay,
		UserID:        userID,
	}

	err = ar.model.SaveSubscription(ftcOrder, c)

	param := alipay.AliPayTradeAppPay{}
	param.NotifyURL = aliNotifyURL
	param.Subject = plan.Description
	param.OutTradeNo = orderID
	param.TotalAmount = plan.GetPriceAli()
	param.ProductCode = aliProductCode
	param.GoodsType = "0"

	values, err := ar.client.URLValues(param)

	if err != nil {
		util.Render(w, util.NewBadRequest(err.Error()))
		return
	}

	util.Render(w, util.NewResponse().SetBody(map[string]string{
		"ftcOrderId": orderID,
		"param":      values.Encode(),
	}))
}

// Notification receives alipay callback
func (ar AliPayRouter) Notification(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()

	if err != nil {
		logger.WithField("location", "AliNotification").Error(err)

		w.Write([]byte("fail"))
		return
	}

	// If err is nil, then the signature is verified.
	noti, err := ar.client.GetTradeNotification(req)

	if err != nil {
		logger.WithField("location", "AliNotification").Error(err)

		w.Write([]byte(fail))
		return
	}

	logger.WithField("location", "AliNotification").Infof("Ali notification data: %+v", noti)

	// 4、验证app_id是否为该商户本身
	if noti.AppId != ar.appID {
		logger.WithField("location", "AliNotification").Info("AppID does not match")

		w.Write([]byte(fail))
		return
	}

	orderID := noti.OutTradeNo

	// 1、商户需要验证该通知数据中的out_trade_no是否为商户系统中创建的订单号
	subs, err := ar.model.FindSubscription(orderID)

	// If the order does not exist, tell ali success;
	// If err is not `not found`, tell ali to resend.
	if err != nil {
		if err == sql.ErrNoRows {
			w.Write([]byte(success))
		} else {
			w.Write([]byte(fail))
		}

		return
	}

	// 2、判断total_amount是否确实为该订单的实际金额（即商户订单创建时的金额）
	if subs.AliTotalAmount() != noti.TotalAmount {
		logger.WithField("location", "AliNotification").Infof("Subscrition total amount: %s vs Notification total amount: %s", subs.AliTotalAmount(), noti.TotalAmount)

		w.Write([]byte(success))

		return
	}

	// 3、校验通知中的seller_id（或者seller_email) 是否为out_trade_no这笔单据的对应的操作方（有的时候，一个商户可能有多个seller_id/seller_email） :-< ???

	// 在上述验证通过后商户必须根据支付宝不同类型的业务通知，正确的进行不同的业务处理，并且过滤重复的通知结果数据。在支付宝的业务通知中，只有交易通知状态为TRADE_SUCCESS或TRADE_FINISHED时，支付宝才会认定为买家付款成功。

	// If this order already confirmed.
	if subs.ConfirmedAt != "" {
		w.Write([]byte(success))

		return
	}

	switch noti.TradeStatus {
	case tradeSuccess, tradeFinished:
		logger.WithField("location", "AliNotification").Infof("Order %s paid", noti.OutTradeNo)

	case tradePending:
		logger.WithField("location", "AliNotification").Info("Payment pending")
		w.Write([]byte(fail))
		return

	case tradeClosed:
		logger.WithField("location", "AliNotification").Info("Transaction closed. Money not earned. :-<")
		w.Write([]byte(success))
		return

	default:
		logger.WithField("location", "AliNotification").Info("Unkknow trade status")
		w.Write([]byte(success))
		return
	}

	confirmTime, err := util.ParseAliTime(noti.GmtPayment)

	if err != nil {
		confirmTime = time.Now()
	}

	err = ar.model.ConfirmSubscription(subs, confirmTime)

	if err != nil {
		w.Write([]byte(fail))
		return
	}

	w.Write([]byte(success))
}

// VerifyAppPay verify the result of native app pay.
// Implements https://docs.open.alipay.com/204/105301/
// 一、同步通知参数说明
func (ar AliPayRouter) VerifyAppPay(w http.ResponseWriter, req *http.Request) {
	var result aliAppPayResult

	if err := parseJSON(req.Body, &result); err != nil {
		util.Render(w, util.NewBadRequest(err.Error()))

		return
	}

	ok, err := ar.client.VerifySign(result.URLValues())

	if err != nil {
		util.Render(w, util.NewBadRequest(err.Error()))

		return
	}

	if !ok {
		util.Render(w, util.NewBadRequest("Verification failed"))

		return
	}

	appID := result.Response["app_id"]
	orderID := result.Response["out_trade_no"]
	totalAmount := result.Response["total_amount"]

	// 4、验证app_id是否为该商户本身
	if appID != ar.appID {
		logger.WithField("location", "AliNotification").Info("AppID does not match")

		util.Render(w, util.NewBadRequest("App ID mismatched"))
		return
	}

	// 1、商户需要验证该通知数据中的out_trade_no是否为商户系统中创建的订单号
	subs, err := ar.model.FindSubscription(orderID)

	// If the order does not exist, tell ali success;
	// If err is not `not found`, tell ali to resend.
	if err != nil {
		util.Render(w, util.NewBadRequest(err.Error()))

		return
	}

	// 2、判断total_amount是否确实为该订单的实际金额（即商户订单创建时的金额）
	if subs.AliTotalAmount() != totalAmount {
		logger.WithField("location", "AliNotification").Infof("Subscrition total amount: %s vs Notification total amount: %s", subs.AliTotalAmount(), totalAmount)

		util.Render(w, util.NewBadRequest("Total amount does not match"))

		return
	}

	util.Render(w, util.NewNoContent())
}
