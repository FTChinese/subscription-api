package controller

import (
	"database/sql"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/objcoding/wxpay"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/wechat"
	"net/http"
	"os"
)

// WxPayRouter wraps wxpay and alipay sdk instances.
type WxPayRouter struct {
	client wechat.Client
	PayRouter
}

// NewWxRouter creates a new instance or OrderRouter
func NewWxRouter(m model.Env, p postoffice.Postman, sandbox bool) WxPayRouter {

	var app wechat.PayApp
	if err := viper.UnmarshalKey("wxapp.m_subs", &app); err != nil {
		logger.WithField("trace", "NewWxRouter").Error(err)
		os.Exit(1)
	}

	// Pay attention to the last parameter.
	// It should always be false because Weixin's sandbox address does not work!
	r := WxPayRouter{
		client: wechat.NewClient(app),
	}
	r.sandbox = sandbox
	r.model = m
	r.postman = p

	return r
}

// UnifiedOrder implements 统一下单.
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_1
func (router WxPayRouter) UnifiedOrder(w http.ResponseWriter, req *http.Request) {
	// Get member tier and billing cycle from url
	tierKey := getURLParam(req, "tier").toString()
	cycleKey := getURLParam(req, "cycle").toString()

	if tierKey == "" || cycleKey == "" {
		view.Render(w, view.NewBadRequest(msgInvalidURI))
		return
	}

	// Try to find a plan based on the tier and cycle.
	plan, err := router.model.GetCurrentPricing().FindPlan(tierKey, cycleKey)
	// If pricing plan if not found.
	if err != nil {
		logger.WithField("trace", "UnifiedOrder").Error(err)

		view.Render(w, view.NewBadRequest(msgInvalidURI))
		return
	}

	logger.WithField("trace", "UnifiedOrder").Infof("Subscritpion plan: %+v", plan)

	// Get user id from request header.
	// If user id is found, it means user is subscribing with FTC account;
	// if union id is found, it means user is subscribing with Wechat account;
	userID := req.Header.Get(userIDKey)
	unionID := req.Header.Get(unionIDKey)
	var loginMethod enum.LoginMethod
	if userID != "" {
		loginMethod = enum.LoginMethodEmail
	} else if unionID != "" {
		loginMethod = enum.LoginMethodWx
		userID = unionID
	}

	subs := paywall.NewWxpaySubs(userID, plan, loginMethod)

	ok, err := router.model.IsSubsAllowed(subs)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}
	if !ok {
		view.Render(w, view.NewForbidden("Already a subscribed user and not within allowed renewal period."))
		return
	}

	// Save this subscription order.
	app := gorest.NewClientApp(req)
	err = router.model.SaveSubscription(subs, app)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// Build Wechat pay parameters.
	param := router.wxUniOrderParam(plan.Description, app.UserIP, subs)

	logger.WithField("trace", "UnifiedOrder").Infof("Unifed order params: %+v", param)

	// Send order to wx
	resp, err := router.client.UnifiedOrder(param)

	if err != nil {
		logger.WithField("trace", "UnifiedOrder").Error(err)

		view.Render(w, view.NewBadRequest(err.Error()))

		return
	}

	// Log the response.
	go func() {
		uor := wechat.NewUnifiedOrderResp(resp)
		router.model.SavePrepayResp(subs.OrderID, uor)
	}()

	if r := router.client.ValidateResponse(resp); r != nil {
		view.Render(w, view.NewUnprocessable(r))
		return
	}

	prepay := router.client.NewPrepay(resp.GetString("prepay_id"), subs)

	sign := router.client.Sign(prepay.Param())

	view.Render(w, view.NewResponse().SetBody(prepay.WithHash(sign)))
}

// Notification implements 支付结果通知
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_7&index=3
func (router WxPayRouter) Notification(w http.ResponseWriter, req *http.Request) {

	resp := wxpay.Notifies{}

	params, err := router.client.ParseResponse(req.Body)

	if err != nil {
		logger.WithField("trace", "Notification").Error(err)

		w.Write([]byte(resp.NotOK(err.Error())))

		return
	}

	// Log the response.
	go func() {
		noti := wechat.NewNotification(params)
		router.model.SaveWxNotification(noti)
	}()

	// Verify appid, mch_id, trade_type, total_fee.
	if r := router.client.ValidateResponse(params); r != nil {
		logger.WithField("trace", "Notification").Error(r.GetMessage())
		w.Write([]byte(resp.OK()))

		return
	}

	// Get out_trade_no to retrieve order.
	// Check the order's confirmed_utc field.
	// If confirmed_utc is empty, get time_end from params and set confirmed_utc to it.
	orderID := params.GetString("out_trade_no")
	subs, err := router.model.FindSubscription(orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.Write([]byte(resp.OK()))
			return
		}
		w.Write([]byte(resp.NotOK(err.Error())))
		return
	}

	charged := params.GetInt64("total_fee")
	if !subs.IsWxChargeMatched(charged) {
		w.Write([]byte(resp.OK()))
		return
	}

	// updatedSubs
	timeEnd := params.GetString("time_end")
	confirmedSubs, err := router.model.ConfirmPayment(orderID, util.ParseWxTime(timeEnd))

	if err != nil {
		logger.WithField("trace", "Notification").Error(err)

		switch err {
		case model.ErrOrderNotFound, model.ErrAlreadyConfirmed:
			w.Write([]byte(resp.OK()))
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
// func (router WxPayRouter) OrderQuery(w http.ResponseWriter, req *http.Request) {
// 	orderID := getURLParam(req, "orderId").toString()

// 	if orderID == "" {
// 		view.Render(w, view.NewBadRequest(""))
// 		return
// 	}

// 	params := make(wxpay.Params)
// 	params.SetString("out_trade_no", orderID)

// 	// Send query to Wechat server
// 	resp, err := router.client.OrderQuery(params)

// 	// If there are any errors when querying order.
// 	if err != nil {
// 		logger.WithField("location", "OrderQuery").Error(err)

// 		view.Render(w, view.NewInternalError(err.Error()))

// 		return
// 	}

// 	logger.WithField("location", "OrderQuery").Infof("Order query result: %+v", resp)

// 	// Reponse fields:
// 	// return_code: SUCCESS|FAIL
// 	// return_msg: string
// 	// appid
// 	// mch_id
// 	// nonce_str
// 	// sign
// 	// result_code
// 	// err_code
// 	// err_code_des
// 	if resp.GetString("return_code") == wxpay.Fail {
// 		returnMsg := resp.GetString("return_msg")
// 		logger.
// 			WithField("location", "OrderQuery").
// 			Errorf("return_code is FAIL. return_msg: %s", returnMsg)

// 		reason := &view.Reason{
// 			Field: "return_code",
// 			Code:  "fail",
// 		}
// 		reason.SetMessage(returnMsg)

// 		view.Render(w, view.NewUnprocessable(reason))

// 		return
// 	}

// 	if resp.GetString("result_code") == wxpay.Fail {
// 		errCode := resp.GetString("err_code")
// 		errCodeDes := resp.GetString("err_code_des")

// 		logger.WithField("location", "OrderQuery").
// 			WithField("err_code", errCode).
// 			WithField("err_code_des", errCodeDes).
// 			Error("Wx unified order result failed")

// 		switch errCode {
// 		case "ORDERNOTEXIST":
// 			view.Render(w, view.NewNotFound())

// 		default:
// 			reason := &view.Reason{
// 				Field: "result_code",
// 				Code:  "fail",
// 			}
// 			reason.SetMessage(errCodeDes)
// 			view.Render(w, view.NewUnprocessable(reason))
// 		}

// 		return
// 	}

// 	if ok := router.client.VerifyIdentity(resp); !ok {
// 		view.Render(w, view.NewNotFound())
// 		return
// 	}

// 	// Response if return_code == SUCCESS and result_code == SUCCESS
// 	// openid
// 	// trade_type: APP
// 	// trade_state: SUCCESS | REFUND | NOTPAY | CLOSED | REVOKED | USERPAYING | PAYERROR
// 	// bank_type
// 	// total_fee
// 	// cash_fee
// 	// transaction_id
// 	// out_trade_no
// 	// time_end: 20091225091010
// 	// trade_state_desc
// 	timeEnd := resp.GetString("time_end")
// 	order := WxQueryOrder{
// 		OpenID:        resp.GetString("openid"),
// 		TradeType:     resp.GetString("trade_type"),
// 		PaymentState:  resp.GetString("trade_state"),
// 		TotalFee:      resp.GetString("total_fee"),
// 		TransactionID: resp.GetString("transaction_id"),
// 		FTCOrderID:    resp.GetString("out_trade_no"),
// 		PaidAt:        timeEnd,
// 	}

// 	view.Render(w, view.NewResponse().SetBody(order))
// }
