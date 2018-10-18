package controller

import (
	"database/sql"
	"fmt"
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
	account := wxpay.NewAccount(wx.AppID, wx.MchID, wx.APIKey, false)

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

// NewWxOrder creates a new order for wxpay.
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

	tier, ok := util.NewTier(tierKey)

	cycle, ok := util.NewCycle(cycleKey)

	if !ok {
		util.Render(w, util.NewBadRequest(msgInvalidURI))
		return
	}

	// Get user id from request header
	userID := req.Header.Get(userIDKey)

	plan, err := util.NewPlan(tier, cycle)

	orderID := util.CreateOrderID(plan)

	if err != nil {
		util.Render(w, util.NewBadRequest(msgInvalidURI))
		return
	}

	// Get request client metadata
	c := util.NewRequestClient(req)

	// Save this order to db.
	ftcOrder := model.SubscribeOrder{
		OrderID:       orderID,
		TierToBuy:     tier,
		BillingCycle:  cycle,
		Price:         plan.Price,
		TotalAmount:   plan.Price * 1,
		PaymentMethod: util.Wxpay,
		UserID:        userID,
	}

	err = wr.model.NewOrder(ftcOrder, c)

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

	// Create prepay order accoding to https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_12&index=2
	appParams := wr.createPrepayOrder(prepayID)

	util.Render(w, util.NewResponse().SetBody(appParams))
}
