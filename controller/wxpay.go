package controller

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/icrowley/fake"

	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/util"
)

// WxPayRouter wraps wxpay and alipay sdk instances.
type WxPayRouter struct {
	config WxConfig
	client *wxpay.Client
	model  model.Env
}

// NewWxRouter creates a new instance or OrderRouter
func NewWxRouter(db *sql.DB, isProd bool) WxPayRouter {
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
		model:  model.Env{DB: db},
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
func (wr WxPayRouter) UnifiedOrder(w http.ResponseWriter, req *http.Request) {
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
	member, err := wr.model.FindMember(userID)

	// If membership for this user is found, and is not in the allowed renewal period.
	// Allowed renewal period: current time is within the length of the expiration time minus the requested billing cycle.
	if err == nil && !member.CanRenew(cycle) {
		util.Render(w, util.NewForbidden("Already a subscribed user and not within allowed renewal period."))
		return
	}

	// Plan if not found.
	plan, err := model.NewPlan(tier, cycle)

	if err != nil {
		logger.WithField("location", "UnifiedOrder").Error(err)

		util.Render(w, util.NewBadRequest(msgInvalidURI))
		return
	}

	if !wr.config.IsProd {
		plan.Price = 0.01
	}

	logger.WithField("location", "UnifiedOrder").Infof("Subscritpion plan: %+v", plan)

	// Order id will be used:
	// 1. Save in our database;
	// 2. Send to wx;
	// 3. Send to app;
	orderID := model.CreateOrderID(plan)

	logger.WithField("location", "UnifiedOrder").Infof("Created order: %s", orderID)

	// Get request client required headers
	c := util.NewRequestClient(req)

	// Save this order to db.
	ftcOrder := model.Subscription{
		OrderID:       orderID,
		TierToBuy:     plan.Tier,
		BillingCycle:  plan.Cycle,
		Price:         plan.Price,
		TotalAmount:   plan.Price,
		PaymentMethod: model.Wxpay,
		UserID:        userID,
	}

	err = wr.model.SaveSubscription(ftcOrder, c)

	if err != nil {
		util.Render(w, util.NewDBFailure(err))
		return
	}
	// Prepare to send wx unified order.
	params := make(wxpay.Params)

	// Compose request parameters
	params.SetString("body", plan.Description).
		SetString("out_trade_no", orderID).
		SetInt64("total_fee", plan.GetPriceCent()).
		SetString("spbill_create_ip", c.UserIP).
		SetString("notify_url", wxNotifyURL).
		SetString("trade_type", "APP")

	if !wr.config.IsProd {
		params.SetString("spbill_create_ip", fake.IPv4())
	}

	logger.WithField("location", "UnifiedOrder").Infof("Order params: %+v", params)

	// Send order to wx
	resp, err := wr.client.UnifiedOrder(params)

	if err != nil {
		logger.WithField("location", "UnifiedOrder").Error(err)

		util.Render(w, util.NewBadRequest(err.Error()))

		return
	}

	// Possible response:
	//  map[return_code:FAIL return_msg:appid不存在]
	logger.WithField("location", "UnifiedOrder").Infof("Wx unified order response: %+v", resp)

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

		reason := &util.Reason{
			Field: "return_code",
			Code:  "fail",
		}
		reason.SetMessage(returnMsg)

		util.Render(w, util.NewUnprocessable(reason))
		return
	}

	if resp.GetString("result_code") == wxpay.Fail {
		errCode := resp.GetString("err_code")
		errCodeDes := resp.GetString("err_code_des")

		logger.WithField("location", "UnifiedOrder").
			WithField("err_code", errCode).
			WithField("err_code_des", errCodeDes).
			Error("Wx unified order result failed")

		reason := &util.Reason{
			Field: "result_code",
			Code:  errCode,
		}
		reason.SetMessage(errCodeDes)
		util.Render(w, util.NewUnprocessable(reason))

		return
	}

	// Get prepay id from response
	prepayID := resp.GetString("prepay_id")

	// Create prepay order according to https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_12&index=2
	appParams := wr.createPrepayOrder(prepayID)
	appParams.SetString("ftcOrderId", orderID)

	util.Render(w, util.NewResponse().SetBody(appParams))
}

func (wr WxPayRouter) createPrepayOrder(prepayID string) wxpay.Params {
	nonce, _ := util.RandomHex(10)

	p := make(wxpay.Params)
	p["appid"] = wr.config.AppID
	p["partnerid"] = wr.config.MchID
	p["prepayid"] = prepayID
	p["package"] = "Sign=WXPay"
	p["noncestr"] = nonce
	p["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	p["sign"] = wr.client.Sign(p)

	return p
}

// Notification implements 支付结果通知
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_7&index=3
func (wr WxPayRouter) Notification(w http.ResponseWriter, req *http.Request) {

	resp := wxpay.Notifies{}

	params, err := wr.processWxResponse(req.Body)

	if err != nil {
		logger.WithField("location", "WxNotification").Error(err)

		w.Write([]byte(resp.NotOK(err.Error())))

		return
	}

	logger.WithField("location", "Wx pay notification").Infof("Wx pay response: %+v", params)
	// Verify appid, mch_id, trade_type, total_fee.
	// Get out_trade_no to retrieve order.
	// Check the order's confirmed_utc field.
	// If confirmed_utc is empty, get time_end from params and set confirmed_utc to it.

	// If this notification does not belong to use, refuse wx's retry.
	if ok := wr.verifyRespIdentity(params); !ok {
		w.Write([]byte(resp.OK()))

		return
	}

	orderID := params.GetString("out_trade_no")
	totalFee := params.GetInt64("total_fee")
	timeEnd := params.GetString("time_end")

	// Find order
	subs, err := wr.model.FindSubscription(orderID)

	// if order is not found
	if err != nil {
		if err == sql.ErrNoRows {
			w.Write([]byte(resp.OK()))
		} else {
			w.Write([]byte(resp.NotOK(err.Error())))
		}
		return
	}

	// Verify total amount
	if totalFee != subs.WxTotalFee() {
		w.Write([]byte(resp.OK()))
		return
	}

	// If order is found, and is already confirmed.
	// Should we check membership data here?
	if subs.ConfirmedAt != "" {
		w.Write([]byte(resp.OK()))
		return
	}

	// For test environment stops here.
	if !wr.config.IsProd {
		w.Write([]byte(resp.OK()))
		return
	}

	// Convert this time end to SQL DATETIME
	// The problem here is we record confirmation time always in UTC. This if fixed.
	confirmTime, err := util.ParseWxTime(timeEnd)

	if err != nil {
		confirmTime = time.Now()
	}

	err = wr.model.ConfirmSubscription(subs, confirmTime)

	if err != nil {
		w.Write([]byte(resp.NotOK(err.Error())))

		return
	}

	w.Write([]byte(resp.OK()))
}

// OrderQuery implements 查询订单
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
// Only transaction_id or out_trade_no is required.
func (wr WxPayRouter) OrderQuery(w http.ResponseWriter, req *http.Request) {
	orderID := getURLParam(req, "orderId").toString()

	if orderID == "" {
		util.Render(w, util.NewBadRequest(""))
		return
	}

	params := make(wxpay.Params)
	params.SetString("out_trade_no", orderID)

	resp, err := wr.client.OrderQuery(params)

	// If there are any errors when querying order.
	if err != nil {
		logger.WithField("location", "OrderQuery").Error(err)

		util.Render(w, util.NewNotFound())

		return
	}

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

		util.Render(w, util.NewBadRequest("Failed to contact wx server"))
		return
	}

	if resp.GetString("result_code") == wxpay.Fail {
		errCode := resp.GetString("err_code")
		errCodeDes := resp.GetString("err_code_des")

		logger.WithField("location", "OrderQuery").
			WithField("err_code", errCode).
			WithField("err_code_des", errCodeDes).
			Error("Wx unified order result failed")
		util.Render(w, util.NewBadRequest(errCodeDes))

		return
	}

	if ok := wr.verifyRespIdentity(resp); !ok {
		util.Render(w, util.NewBadRequest("Wrong identity"))
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
	order := WxOrder{
		OpenID:        resp.GetString("openid"),
		TradeType:     resp.GetString("trade_type"),
		PaymentState:  resp.GetString("trade_state"),
		TotalFee:      resp.GetString("total_fee"),
		TransactionID: resp.GetString("transaction_id"),
		FTCOrderID:    resp.GetString("out_trade_no"),
		PaidAt:        util.ISO8601UTC.FromWx(timeEnd),
	}

	util.Render(w, util.NewResponse().SetBody(order))
}

func (wr WxPayRouter) processWxResponse(r io.Reader) (wxpay.Params, error) {

	var returnCode string
	params := util.Decode(r)
	logger.WithField("location", "processWxResponse").Infof("Resp params: %+v", params)

	if params.ContainsKey("return_code") {
		returnCode = params.GetString("return_code")
	} else {
		return nil, errors.New("no return_code in XML")
	}

	switch returnCode {
	case wxpay.Fail:
		return nil, errors.New("wx notification failed")

	case wxpay.Success:
		if wr.client.ValidSign(params) {
			logger.WithField("location", "process wx response").Info("Validating signature passed")
			return params, nil
		}
		return nil, errors.New("invalid sign value in XML")

	default:
		return nil, errors.New("return_code value is invalid in XML")
	}
}

func (wr WxPayRouter) verifyRespIdentity(params wxpay.Params) bool {
	if params.ContainsKey("appid") {
		logger.WithField("location", "Verify wx response id").Error("Missing appid")
		return false
	}

	if params.ContainsKey("mch_id") {
		logger.WithField("location", "Verify wx response id").Error("Missing mch_id")
		return false
	}

	if params.GetString("appid") != wr.config.AppID {
		logger.WithField("location", "Verify wx response id").Error("appid does not match")
		return false
	}

	if params.GetString("mch_id") != wr.config.MchID {
		logger.WithField("location", "Verify wx response id").Error("mch_id does not match")
		return false
	}

	return true
}
