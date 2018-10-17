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

// OrderRouter wraps wxpay and alipay sdk instances.
type OrderRouter struct {
	wxConfig util.WxConfig
	wxClient *wxpay.Client
	model    model.Env
}

// NewOrderRouter creates a new instance or OrderRouter
func NewOrderRouter(wx util.WxConfig, db *sql.DB) OrderRouter {
	account := wxpay.NewAccount(wx.AppID, wx.MchID, wx.APIKey, false)

	return OrderRouter{
		model:    model.Env{DB: db},
		wxConfig: wx,
		wxClient: wxpay.NewClient(account),
	}
}

func (o OrderRouter) createPrepayOrder(prepayID string) wxpay.Params {
	p := make(wxpay.Params)
	p["appid"] = o.wxConfig.AppID
	p["partnerid"] = o.wxConfig.MchID
	p["prepayid"] = prepayID
	p["package"] = "Sign=WXPay"
	p["noncestr"] = ""
	p["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	p["sign"] = o.wxClient.Sign(p)

	return p
}

// NewWxOrder creates a new order for wxpay
func (o OrderRouter) NewWxOrder(w http.ResponseWriter, req *http.Request) {
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

	err = o.model.NewOrder(ftcOrder, c)

	// Prepare to send wx unified order.
	params := make(wxpay.Params)

	params.SetString("body", plan.Description).
		SetString("out_trade_no", orderID).
		SetInt64("total_fee", plan.GetPriceCent()).
		SetString("spbill_create_ip", c.UserIP).
		SetString("notify_url", wxNotifyURL).
		SetString("trade_type", "APP")

	// Send order to wx
	resp, err := o.wxClient.UnifiedOrder(params)

	if err != nil {
		util.Render(w, util.NewInternalError(err.Error()))

		return
	}

	// Get prepay id from response
	prepayID := resp.GetString("prepay_id")

	// Create prepay order accoding to https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_12&index=2
	appParams := o.createPrepayOrder(prepayID)

	util.Render(w, util.NewResponse().SetBody(appParams))
}
