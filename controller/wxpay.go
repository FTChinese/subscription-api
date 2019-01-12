package controller

import (
	"net/http"
	"os"

	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/view"
	"gitlab.com/ftchinese/subscription-api/wechat"
)

// WxPayRouter wraps wxpay and alipay sdk instances.
type WxPayRouter struct {
	client wechat.Client
	model  model.Env
}

// NewWxRouter creates a new instance or OrderRouter
func NewWxRouter(env model.Env) WxPayRouter {
	appID := os.Getenv("WXPAY_APPID")
	mchID := os.Getenv("WXPAY_MCHID")
	apiKey := os.Getenv("WXPAY_API_KEY")

	// Pay attention to the last parameter.
	// It should always be false because Weixin's sandbox address does not work!

	return WxPayRouter{
		client: wechat.NewClient(appID, mchID, apiKey),
		model:  env,
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
		loginMethod = enum.EmailLogin
	} else if unionID != "" {
		loginMethod = enum.WechatLogin
		userID = unionID
	}

	subs := model.NewWxpaySubs(userID, plan, loginMethod)

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
		SetInt64("total_fee", plan.PriceForWx()).
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

	if ok := router.client.VerifyIdentity(resp); !ok {
		view.Render(w, view.NewBadRequest("Wechat response contains mismatching IDs"))
		return
	}

	if r := wechat.ValidateResponse(resp); r != nil {
		view.Render(w, view.NewUnprocessable(r))
	}

	order := router.client.BuildPrepayOrder(subs.OrderID, subs.Price, resp.GetString("prepay_id"))

	view.Render(w, view.NewResponse().SetBody(order))
}

// Notification implements 支付结果通知
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_7&index=3
func (router WxPayRouter) Notification(w http.ResponseWriter, req *http.Request) {

	resp := wxpay.Notifies{}

	params, err := router.client.ParseResponse(req.Body)

	if err != nil {
		w.Write([]byte(resp.NotOK(err.Error())))

		return
	}

	// Verify appid, mch_id, trade_type, total_fee.
	if ok := router.client.VerifyIdentity(params); !ok {
		w.Write([]byte(resp.OK()))

		return
	}

	if err := router.model.VerifyWxNotification(params); err != nil {
		w.Write([]byte(resp.OK()))
	}

	// For logging purpose.
	go func() {
		n := wechat.NewNotification(params)
		router.model.SaveWxNotification(n)
	}()

	// Get out_trade_no to retrieve order.
	// Check the order's confirmed_utc field.
	// If confirmed_utc is empty, get time_end from params and set confirmed_utc to it.
	orderID := params.GetString("out_trade_no")
	timeEnd := params.GetString("time_end")

	// updatedSubs,
	cofirmedSubs, err := router.model.ConfirmPayment(orderID, util.ParseWxTime(timeEnd))

	if err != nil {
		w.Write([]byte(resp.NotOK(err.Error())))

		return
	}

	// Send a letter to this user.
	go router.model.SendConfirmationLetter(cofirmedSubs)

	w.Write([]byte(resp.OK()))
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

	if ok := router.client.VerifyIdentity(resp); !ok {
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
		PaidAt:        timeEnd,
	}

	view.Render(w, view.NewResponse().SetBody(order))
}
