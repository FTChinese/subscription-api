package controller

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/view"
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
func NewAliRouter(m model.Env, isProd bool) AliPayRouter {
	appID := os.Getenv("ALIPAY_APP_ID")

	// Ali's public key is used to verify alipay's response.
	publicKey, err := ioutil.ReadFile("alipay_public_key.pem")
	if err != nil {
		logger.WithField("location", "NewAliRouter").Error(err)
		os.Exit(1)
	}

	// Private key is used to sign our data that will be sent to alipay.
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
		model:  m,
	}
}

// AppOrder an alipay order for native app.
func (ar AliPayRouter) AppOrder(w http.ResponseWriter, req *http.Request) {
	// Get member tier and billing cycle from url
	tierKey := getURLParam(req, "tier").toString()
	cycleKey := getURLParam(req, "cycle").toString()

	if tierKey == "" || cycleKey == "" {
		view.Render(w, view.NewBadRequest(msgInvalidURI))
		return
	}

	// Get user id from request header
	userID := req.Header.Get(userIDKey)

	plan, err := ar.model.FindPlan(tierKey, cycleKey)

	if err != nil {
		logger.WithField("location", "AliAppOrder").Error(err)

		view.Render(w, view.NewBadRequest(msgInvalidURI))
		return
	}

	logger.WithField("location", "AliAppOrder").Infof("Subscritpion plan: %+v", plan)

	subs := plan.CreateSubs(userID, enum.Alipay)

	// Get request client metadata
	c := util.NewRequestClient(req)

	err = ar.model.PlaceOrder(subs, c)

	if err != nil {
		if err == util.ErrRenewalForbidden {
			view.Render(w, view.NewForbidden("Already a subscribed user and not within allowed renewal period."))
			return
		}

		view.Render(w, view.NewDBFailure(err))
		return
	}

	param := alipay.AliPayTradeAppPay{}
	param.NotifyURL = aliNotifyURL
	param.Subject = plan.Description
	param.OutTradeNo = subs.OrderID
	param.TotalAmount = plan.GetPriceString()
	param.ProductCode = aliProductCode
	param.GoodsType = "0"

	// Call URLValues to generate alipay required data structure and sign it.
	values, err := ar.client.URLValues(param)

	logger.WithField("location", "AliAppOrder").Infof("App pay param: %+v\n", values)

	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	order := AliOrder{
		FtcOrderID: subs.OrderID,
		Price:      plan.Price,
		Param:      values.Encode(),
	}

	view.Render(w, view.NewResponse().SetBody(order))
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
			logger.WithField("location", "AliNotification").Info("Suscription order is not found")
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
	if !subs.ConfirmedAt.IsEmpty() {
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

	updatedSubs, err := ar.model.ConfirmSubscription(subs, confirmTime)

	if err != nil {
		w.Write([]byte(fail))
		return
	}

	err = ar.model.CreateOrUpdateMember(updatedSubs)

	if err != nil {
		w.Write([]byte(fail))
		return
	}

	go ar.model.SendConfirmationLetter(updatedSubs)

	w.Write([]byte(success))
}

// VerifyAppPay verify the result of native app pay.
// Implements https://docs.open.alipay.com/204/105301/
// 一、同步通知参数说明
func (ar AliPayRouter) VerifyAppPay(w http.ResponseWriter, req *http.Request) {
	var result aliAppPayResult

	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		view.Render(w, view.NewBadRequest("Problems parsing JSON"))

		return
	}

	logger.WithField("location", "VerifyAppPay").Infof("Ali app pay result: %s", body)

	signedStr := extractAppPayResp(string(body), keyAppPayResp)

	if err := json.Unmarshal(body, &result); err != nil {
		logger.WithField("location", "VerifyAppPay").Error(err)
		view.Render(w, view.NewBadRequest(err.Error()))

		return
	}

	sign := result.Sign
	signType := result.SignType
	ok, err := verifyAliResp([]byte(signedStr), sign, signType, ar.client.AliPayPublicKey)

	// 422
	if err != nil {
		reason := &view.Reason{
			Field: "sign",
			Code:  util.CodeInvalid,
		}
		reason.SetMessage(err.Error())

		view.Render(w, view.NewUnprocessable(reason))

		return
	}

	if !ok {
		reason := &view.Reason{
			Field: "sign",
			Code:  util.CodeIncorrect,
		}
		reason.SetMessage(err.Error())

		view.Render(w, view.NewUnprocessable(reason))

		return
	}

	appID := result.Response.AppID
	orderID := result.Response.FtcOrderID
	totalAmount := result.Response.TotalAmount

	// 4、验证app_id是否为该商户本身
	if appID != ar.appID {
		logger.WithField("location", "AliNotification").Info("AppID does not match")

		reason := &view.Reason{
			Field: "app_id",
			Code:  util.CodeIncorrect,
		}
		reason.SetMessage("APP ID mismatched.")

		view.Render(w, view.NewUnprocessable(reason))
		return
	}

	// 1、商户需要验证该通知数据中的out_trade_no是否为商户系统中创建的订单号
	subs, err := ar.model.FindSubscription(orderID)

	// If the order does not exist, tell ali success;
	// If err is not `not found`, tell ali to resend.
	if err != nil {
		view.Render(w, view.NewDBFailure(err))

		return
	}

	// 2、判断total_amount是否确实为该订单的实际金额（即商户订单创建时的金额）
	if subs.AliTotalAmount() != totalAmount {
		logger.
			WithField("location", "AliNotification").
			Infof("Subscrition total amount: %s vs Notification total amount: %s", subs.AliTotalAmount(), totalAmount)

		reason := &view.Reason{
			Field: "total_amount",
			Code:  util.CodeIncorrect,
		}
		reason.SetMessage("Total amount does not match.")
		view.Render(w, view.NewUnprocessable(reason))

		return
	}

	paidAt, _ := util.ToISO8601UTC.
		FromDatetime(
			result.Response.DateTime,
			util.TZShanghai,
		)

	order := AliAppOrder{
		FtcOrderID: result.Response.FtcOrderID,
		AliOrderID: result.Response.AliOrderID,
		PaidAt:     paidAt,
	}

	view.Render(w, view.NewResponse().SetBody(order))
}
