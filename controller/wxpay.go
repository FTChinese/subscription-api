package controller

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/view"
)

// WxPayRouter wraps wxpay and alipay sdk instances.
type WxPayRouter struct {
	config WxConfig
	client *wxpay.Client
	model  model.Env
}

// NewWxRouter creates a new instance or OrderRouter
func NewWxRouter(m model.Env, isProd bool) WxPayRouter {
	config := WxConfig{
		AppID:  os.Getenv("WXPAY_APPID"),
		MchID:  os.Getenv("WXPAY_MCHID"),
		APIKey: os.Getenv("WXPAY_API_KEY"),
		IsProd: isProd,
	}
	// Pay attention to the last parameter.
	// It should always be false because Weixin's sandbox address does not work!
	account := wxpay.NewAccount(config.AppID, config.MchID, config.APIKey, false)

	return WxPayRouter{
		model:  m,
		config: config,
		client: wxpay.NewClient(account),
	}
}

// UnifiedOrder implements 统一下单.
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_1
//
// Workflow
//
// 1. First extract userID from request header;
//
// 2. Use this userID to retrieve this user's membership from `premium.ftc_vip` table;
//
// 3. If the membership is not found, proceed to order creation;
//
// 4. If the membership is found, compare the membership's expiration date:
// If the difference between now and expire_time is less than billing_cycle, then the user is allowed to subscribe to next billing cycle;
// If the difference between now and expire_time is greater than billing_cycle, it means user has already pre-subscribed to next billing cycle, refuse this request.
//
// 5. Generate order.
//
// After order is confirmed by payment providers:
//
// 1. Use response message to find order, and then find the userID;
//
// 2. Try to find if membership for this userID exists;
//
// 3. If not exists, this is a new subscription, simply add a new record with start_utc and expire_utc columns set to current time and current time + billing cycle respectively;
//
// 4. If the membership already exists, then check whether expire_utc is before now;
//
// 5. If the expire_utc is before now, it means this user's membership has already expired, he is re-subscribing now, so treat it as a new subscription: update member_tier, billing_cycle, start_utc and expire_utc;
//
// 6. If the expire_utc is after now, it means this user is renewing subscription, the expire_utc should be the the current value + next billing cycle. `start_utc` remain unchanged.
func (router WxPayRouter) UnifiedOrder(w http.ResponseWriter, req *http.Request) {
	// Get member tier and billing cycle from url
	tierKey := getURLParam(req, "tier").toString()
	cycleKey := getURLParam(req, "cycle").toString()

	if tierKey == "" || cycleKey == "" {
		view.Render(w, view.NewBadRequest(msgInvalidURI))
		return
	}

	// Try to find a plan based on the tier and cycle.
	plan, err := router.model.LoadCurrentPlans().FindPlan(tierKey, cycleKey)
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
		loginMethod = enum.EmailLogin
	} else if unionID != "" {
		loginMethod = enum.WechatLogin
		userID = unionID
	}

	subs := model.NewWxSubs(userID, plan, loginMethod)

	ok, err := router.model.IsSubsAllowed(subs)
	// err = wr.model.PlaceOrder(subs, app)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}
	if !ok {
		view.Render(w, view.NewForbidden("Already a subscribed user and not within allowed renewal period."))
		return
	}

	// Save this subscription order.
	app := util.NewClientApp(req)
	err = router.model.SaveSubscription(subs, app)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}
	// Prepare to send wx unified order.
	params := make(wxpay.Params)

	// Compose request parameters
	params.SetString("body", plan.Description).
		SetString("out_trade_no", subs.OrderID).
		SetInt64("total_fee", plan.GetPriceCent()).
		SetString("spbill_create_ip", app.UserIP).
		SetString("notify_url", wxNotifyURL).
		SetString("trade_type", "APP")

	logger.WithField("trace", "UnifiedOrder").Infof("Order params: %+v", params)

	// Send order to wx
	resp, err := router.client.UnifiedOrder(params)

	if err != nil {
		logger.WithField("trace", "UnifiedOrder").Error(err)

		view.Render(w, view.NewBadRequest(err.Error()))

		return
	}

	// Possible response:
	//  map[return_code:FAIL return_msg:appid不存在]
	logger.WithField("trace", "UnifiedOrder").Infof("Wx unified order response: %+v", resp)

	// Example response:
	// return_code:SUCCESS|FAIL
	// return_msg:OK
	//
	// Present only if return_code == SUCCESS
	// appid:wx......
	// mch_id:........
	// nonce_str:8p8ZlUBkLsFPxC6g
	// sign:DB68F0D9F193D499DF9A2EDBFFEAF312
	// result_code:SUCCESS|FAIL
	// err_code
	// err_code_des
	//
	// Present only if returnd_code == SUCCESS and result_code == SUCCCESS
	// trade_type:APP
	// prepay_id:wx20125006086590be8d9519f40090763508

	// NOTE: this sdk treat return_code == FAIL as valid.
	// Possible return_msg:
	// appid不存在;
	// 商户号mch_id与appid不匹配;
	// invalid spbill_create_ip;
	// spbill_create_ip参数长度有误; (Wx does not accept IPv6 like 9b5b:2ef9:6c9f:cf5:130e:984d:8958:75f9 :-<)
	if resp.GetString("return_code") == wxpay.Fail {
		returnMsg := resp.GetString("return_msg")
		logger.
			WithField("location", "UnifiedOrder").
			Errorf("return_code is FAIL. return_msg: %s", returnMsg)

		reason := &view.Reason{
			Field: "return_code",
			Code:  "fail",
		}
		reason.SetMessage(returnMsg)

		view.Render(w, view.NewUnprocessable(reason))
		return
	}

	if resp.GetString("result_code") == wxpay.Fail {
		errCode := resp.GetString("err_code")
		errCodeDes := resp.GetString("err_code_des")

		logger.WithField("location", "UnifiedOrder").
			WithField("err_code", errCode).
			WithField("err_code_des", errCodeDes).
			Error("Wx unified order result failed")

		reason := &view.Reason{
			Field: "result_code",
			Code:  errCode,
		}
		reason.SetMessage(errCodeDes)
		view.Render(w, view.NewUnprocessable(reason))

		return
	}

	// Get prepay id from response
	prepayID := resp.GetString("prepay_id")

	// Create prepay order according to https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_12&index=2
	// appParams := wr.createPrepayOrder(prepayID)
	// appParams.SetString("ftcOrderId", orderID)

	nonce, _ := util.RandomHex(10)

	order := WxOrder{
		FtcOrderID: subs.OrderID,
		Price:      plan.Price,
		AppID:      router.config.AppID,
		PartnerID:  router.config.MchID,
		PrepayID:   prepayID,
		Package:    "Sign=WXPay",
		Nonce:      nonce,
		Timestamp:  fmt.Sprintf("%d", time.Now().Unix()),
	}

	order.Signature = router.signOrder(order)

	// appParams.SetString("price", plan.GetPriceString())

	view.Render(w, view.NewResponse().SetBody(order))
}

func (router WxPayRouter) signOrder(order WxOrder) string {
	p := make(wxpay.Params)
	p["appid"] = order.AppID
	p["partnerid"] = order.PartnerID
	p["prepayid"] = order.PrepayID
	p["package"] = order.Package
	p["noncestr"] = order.Nonce
	p["timestamp"] = order.Timestamp

	return router.client.Sign(p)
}

// Notification implements 支付结果通知
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_7&index=3
func (router WxPayRouter) Notification(w http.ResponseWriter, req *http.Request) {

	resp := wxpay.Notifies{}

	params, err := router.processWxResponse(req.Body)

	if err != nil {
		logger.WithField("trace", "WxNotification").Error(err)

		w.Write([]byte(resp.NotOK(err.Error())))

		return
	}

	logger.WithField("location", "WxPayNotification").Infof("Successfully received wechat notification")

	// Verify appid, mch_id, trade_type, total_fee.
	// Get out_trade_no to retrieve order.
	// Check the order's confirmed_utc field.
	// If confirmed_utc is empty, get time_end from params and set confirmed_utc to it.

	// If this notification does not belong to use, refuse wx's retry.
	if ok := router.verifyRespIdentity(params); !ok {
		logger.WithField("location", "WxPayNotification").Info("App idnetity verification not passed")

		w.Write([]byte(resp.OK()))

		return
	}

	orderID := params.GetString("out_trade_no")
	totalFee := params.GetInt64("total_fee")
	timeEnd := params.GetString("time_end")

	// Find order
	subs, err := router.model.FindSubscription(orderID)

	// if order is not found
	if err != nil {
		if err == sql.ErrNoRows {
			logger.WithField("location", "WxPayNotification").Infof("Order %s is not found", orderID)

			w.Write([]byte(resp.OK()))
		} else {
			w.Write([]byte(resp.NotOK(err.Error())))
		}
		return
	}

	// Verify total amount
	if totalFee != subs.WxTotalFee() {
		logger.WithField("location", "WxPayNotification").Infof("Total fee does not match. Should be %d, actual %d", subs.WxTotalFee(), totalFee)

		w.Write([]byte(resp.OK()))
		return
	}

	// If order is found, and is already confirmed.
	// Should we check membership data here?
	if !subs.ConfirmedAt.IsZero() {
		logger.WithField("location", "WxPayNotification").Infof("Order %s already confirmed.", orderID)

		w.Write([]byte(resp.OK()))
		return
	}

	// Convert this time end to SQL DATETIME
	// The problem here is we record confirmation time always in UTC. This if fixed.
	confirmTime, err := util.ParseWxTime(timeEnd)

	if err != nil {
		confirmTime = time.Now()
	}

	updatedSubs, err := router.model.ConfirmSubscription(subs, confirmTime)

	if err != nil {
		w.Write([]byte(resp.NotOK(err.Error())))

		return
	}

	err = router.model.CreateMembership(updatedSubs)

	if err != nil {
		w.Write([]byte(resp.NotOK(err.Error())))

		return
	}

	// Send a letter to this user.
	go router.model.SendConfirmationLetter(subs)

	w.Write([]byte(resp.OK()))
}

func (router WxPayRouter) processWxResponse(r io.Reader) (wxpay.Params, error) {

	var returnCode string
	params := util.Decode(r)
	logger.WithField("trace", "processWxResponse").Infof("Resp params: %+v", params)

	if params.ContainsKey("return_code") {
		returnCode = params.GetString("return_code")
	} else {
		return nil, errors.New("no return_code in XML")
	}

	switch returnCode {
	case wxpay.Fail:
		return nil, errors.New("wx notification failed")

	case wxpay.Success:
		if router.client.ValidSign(params) {
			logger.WithField("location", "processWxResponse").Info("Validating signature passed")
			return params, nil
		}
		return nil, errors.New("invalid sign value in XML")

	default:
		return nil, errors.New("return_code value is invalid in XML")
	}
}

// OrderQuery implements 查询订单
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
// Only transaction_id or out_trade_no is required.
func (router WxPayRouter) OrderQuery(w http.ResponseWriter, req *http.Request) {
	orderID := getURLParam(req, "orderId").toString()

	if orderID == "" {
		view.Render(w, view.NewBadRequest(""))
		return
	}

	params := make(wxpay.Params)
	params.SetString("out_trade_no", orderID)

	// Send query to Wechat server
	resp, err := router.client.OrderQuery(params)

	// If there are any errors when querying order.
	if err != nil {
		logger.WithField("location", "OrderQuery").Error(err)

		view.Render(w, view.NewInternalError(err.Error()))

		return
	}

	logger.WithField("location", "OrderQuery").Infof("Order query result: %+v", resp)

	// Reponse fields:
	// return_code: SUCCESS|FAIL
	// return_msg: string
	// appid
	// mch_id
	// nonce_str
	// sign
	// result_code
	// err_code
	// err_code_des
	if resp.GetString("return_code") == wxpay.Fail {
		returnMsg := resp.GetString("return_msg")
		logger.
			WithField("location", "OrderQuery").
			Errorf("return_code is FAIL. return_msg: %s", returnMsg)

		reason := &view.Reason{
			Field: "return_code",
			Code:  "fail",
		}
		reason.SetMessage(returnMsg)

		view.Render(w, view.NewUnprocessable(reason))

		return
	}

	if resp.GetString("result_code") == wxpay.Fail {
		errCode := resp.GetString("err_code")
		errCodeDes := resp.GetString("err_code_des")

		logger.WithField("location", "OrderQuery").
			WithField("err_code", errCode).
			WithField("err_code_des", errCodeDes).
			Error("Wx unified order result failed")

		switch errCode {
		case "ORDERNOTEXIST":
			view.Render(w, view.NewNotFound())

		default:
			reason := &view.Reason{
				Field: "result_code",
				Code:  "fail",
			}
			reason.SetMessage(errCodeDes)
			view.Render(w, view.NewUnprocessable(reason))
		}

		return
	}

	if ok := router.verifyRespIdentity(resp); !ok {
		logger.WithField("location", "OrderQuery").Info("appid or mch_id mismatched")

		view.Render(w, view.NewNotFound())
		return
	}

	// Response if return_code == SUCCESS and result_code == SUCCESS
	// openid
	// trade_type: APP
	// trade_state: SUCCESS | REFUND | NOTPAY | CLOSED | REVOKED | USERPAYING | PAYERROR
	// bank_type
	// total_fee
	// cash_fee
	// transaction_id
	// out_trade_no
	// time_end: 20091225091010
	// trade_state_desc
	timeEnd := resp.GetString("time_end")
	order := WxQueryOrder{
		OpenID:        resp.GetString("openid"),
		TradeType:     resp.GetString("trade_type"),
		PaymentState:  resp.GetString("trade_state"),
		TotalFee:      resp.GetString("total_fee"),
		TransactionID: resp.GetString("transaction_id"),
		FTCOrderID:    resp.GetString("out_trade_no"),
		PaidAt:        util.ToISO8601UTC.FromWx(timeEnd),
	}

	view.Render(w, view.NewResponse().SetBody(order))
}

func (router WxPayRouter) verifyRespIdentity(params wxpay.Params) bool {
	if !params.ContainsKey("appid") {
		logger.WithField("location", "WxPay verifyRespIdentity").Error("Missing appid")
		return false
	}

	if !params.ContainsKey("mch_id") {
		logger.WithField("location", "WxPay verifyRespIdentity").Error("Missing mch_id")
		return false
	}

	if params.GetString("appid") != router.config.AppID {
		logger.WithField("location", "WxPay verifyRespIdentity").Error("appid does not match")
		return false
	}

	if params.GetString("mch_id") != router.config.MchID {
		logger.WithField("location", "WxPay verifyRespIdentity").Error("mch_id does not match")
		return false
	}

	return true
}
