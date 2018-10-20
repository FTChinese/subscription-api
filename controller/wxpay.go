package controller

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/util"
)

// WxPayRouter wraps wxpay and alipay sdk instances.
type WxPayRouter struct {
	wxConfig util.WxConfig
	wxClient *wxpay.Client
	model    model.Env
}

// NewOrderRouter creates a new instance or OrderRouter
func NewOrderRouter(wx util.WxConfig, db *sql.DB) WxPayRouter {
	account := wxpay.NewAccount(wx.AppID, wx.MchID, wx.APIKey, wx.IsSandbox)

	return WxPayRouter{
		model:    model.Env{DB: db},
		wxConfig: wx,
		wxClient: wxpay.NewClient(account),
	}
}

func (wr WxPayRouter) createPrepayOrder(prepayID string) wxpay.Params {
	nonce, _ := util.RandomHex(10)

	p := make(wxpay.Params)
	p["appid"] = wr.wxConfig.AppID
	p["partnerid"] = wr.wxConfig.MchID
	p["prepayid"] = prepayID
	p["package"] = "Sign=WXPay"
	p["noncestr"] = nonce
	p["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	p["sign"] = wr.wxClient.Sign(p)

	return p
}

// NewWxOrder implements 统一下单.
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
func (wr WxPayRouter) NewWxOrder(w http.ResponseWriter, req *http.Request) {
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
		reason := util.InvalidReason{
			Message: "Already a subscribed user",
			Field:   "order",
			Code:    util.CodeAlreadyExsits,
		}
		util.Render(w, util.NewUnprocessable(reason))
		return
	}

	plan, err := model.NewPlan(tier, cycle)

	orderID := model.CreateOrderID(plan)

	if err != nil {
		util.Render(w, util.NewBadRequest(msgInvalidURI))
		return
	}

	// Get request client metadata
	c := util.NewRequestClient(req)

	// Save this order to db.
	ftcOrder := model.Subscription{
		OrderID:       orderID,
		TierToBuy:     plan.Tier,
		BillingCycle:  plan.Cycle,
		Price:         float32(plan.Price),
		TotalAmount:   float32(plan.Price),
		PaymentMethod: model.Wxpay,
		UserID:        userID,
	}

	err = wr.model.NewSubscription(ftcOrder, c)

	// Prepare to send wx unified order.
	params := make(wxpay.Params)

	params.SetString("body", plan.Description).
		SetString("out_trade_no", orderID).
		SetInt64("total_fee", plan.GetPriceCent()).
		SetString("spbill_create_ip", c.UserIP).
		SetString("notify_url", wxNotifyURL).
		SetString("trade_type", "APP")

	// Send order to wx
	resp, err := wr.wxClient.UnifiedOrder(params)

	if err != nil {
		util.Render(w, util.NewInternalError(err.Error()))

		return
	}

	// Get prepay id from response
	prepayID := resp.GetString("prepay_id")

	// Create prepay order according to https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_12&index=2
	appParams := wr.createPrepayOrder(prepayID)

	util.Render(w, util.NewResponse().SetBody(appParams))
}

// Notification implements 支付结果通知
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_7&index=3
func (wr WxPayRouter) Notification(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)

	// Reply to wx.
	if err != nil {
		resp := buildWxReply("Cannot parse request body", false)
		w.Write(resp)

		return
	}

	params, err := wr.processWxResponse(string(body))

	if err != nil {
		logger.WithField("location", "Wx pay notification").Error(err)

		w.Write(buildWxReply(err.Error(), false))

		return
	}

	logger.WithField("location", "Wx pay notification").Infof("Wx pay response: %+v", params)
	// Verify appid, mch_id, trade_type, total_fee.
	// Get out_trade_no to retrieve order.
	// Check the order's confirmed_utc field.
	// If confirmed_utc is empty, get time_end from params and set confirmed_utc to it.

	if ok := wr.verifyRespIdentity(params); !ok {
		w.Write(buildWxReply("", true))

		return
	}

	orderID := params.GetString("out_trade_no")
	totalFee := params.GetInt64("total_fee")
	timeEnd := params.GetString("time_end")

	// Find order
	subs, err := wr.model.FindSubscription(orderID)

	// if order is not found
	if err != nil {
		w.Write(buildWxReply("", true))
		return
	}

	// Verify total amount
	if totalFee != subs.WxTotalFee() {
		w.Write(buildWxReply("", true))
		return
	}

	// If order is found, and is already confirmed.
	// Should we check membership data here?
	if subs.ConfirmedAt != "" {
		w.Write(buildWxReply("", true))
		return
	}

	// Convert this time end to SQL DATETIME
	// The problem here is we record confirmation time always in UTC. This if fixed.
	confirmTime := util.ParseWxTime(timeEnd)

	// For sandbox environment stop here.
	if wr.wxConfig.IsSandbox {
		w.Write(buildWxReply("", true))
		return
	}

	err = wr.model.ConfirmSubscription(subs, confirmTime)

	if err != nil {
		w.Write(buildWxReply(err.Error(), false))

		return
	}

	w.Write(buildWxReply("", true))
}

func (wr WxPayRouter) processWxResponse(xmlStr string) (wxpay.Params, error) {
	logger.WithField("location", "process wx response").Info(xmlStr)

	var returnCode string
	params := wxpay.XmlToMap(xmlStr)
	if params.ContainsKey("return_code") {
		returnCode = params.GetString("return_code")

		logger.WithField("location", "process wx response").Infof("Wx return_code: %s", returnCode)
	} else {
		return nil, errors.New("no return_code in XML")
	}

	switch returnCode {
	case wxpay.Fail:
		return nil, errors.New("wx notification failed")

	case wxpay.Success:
		if wr.wxClient.ValidSign(params) {
			logger.WithField("location", "process wx response").Info("Valiating signature passed")
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

	if params.GetString("appid") != wr.wxConfig.AppID {
		logger.WithField("location", "Verify wx response id").Error("appid does not match")
		return false
	}

	if params.GetString("mch_id") != wr.wxConfig.MchID {
		logger.WithField("location", "Verify wx response id").Error("mch_id does not match")
		return false
	}

	return true
}

func buildWxReply(msg string, isSuccess bool) []byte {
	p := make(wxpay.Params)
	if isSuccess {
		p["return_code"] = wxpay.Success
		p["return_msg"] = "OK"
	} else {
		p["return_code"] = wxpay.Fail
		p["return_msg"] = msg
	}

	xmlStr := wxpay.MapToXml(p)

	return []byte(xmlStr)
}
