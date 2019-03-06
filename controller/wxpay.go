package controller

import (
	"database/sql"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/wechat"
	"net/http"
	"os"
	"time"
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

	tier, err := GetURLParam(req, "tier").ToString()
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	cycle, err := GetURLParam(req, "cycle").ToString()
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// Try to find a plan based on the tier and cycle.
	plan, err := router.model.GetCurrentPricing().FindPlan(tier, cycle)

	// If pricing plan is not found.
	if err != nil {
		logger.WithField("trace", "UnifiedOrder").Error(err)
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	logger.WithField("trace", "UnifiedOrder").Infof("Subscription plan: %+v", plan)

	// Get user id from request header.
	// If user id is found, it means user is subscribing with FTC account;
	// if union id is found, it means user is subscribing with Wechat account;
	uID := req.Header.Get(userIDKey)
	wID := req.Header.Get(unionIDKey)

	userID := null.NewString(uID, uID != "")
	unionID := null.NewString(wID, wID != "")

	subs, err := paywall.NewWxpaySubs(userID, unionID, plan)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

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

	confirmedAt, err := util.ParseWxTime(timeEnd)
	if err != nil {
		confirmedAt = time.Now()
	}
	confirmedSubs, err := router.model.ConfirmPayment(orderID, confirmedAt)

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
func (router WxPayRouter) OrderQuery(w http.ResponseWriter, req *http.Request) {
	orderID, err := GetURLParam(req, "orderId").ToString()

	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	params := make(wxpay.Params)
	params.SetString("out_trade_no", orderID)

	// Send query to Wechat server
	// Returns the parsed response as a map.
	// It checks if the response contains `return_code` key.
	// If return_code == FAIL, it does not returns error.
	// If return_code == SUCCESS, it verifies the signature.
	resp, err := router.client.OrderQuery(params)

	// If there are any errors when querying order.
	if err != nil {
		logger.WithField("trace", "OrderQuery").Error(err)

		view.Render(w, view.NewInternalError(err.Error()))

		return
	}

	logger.WithField("trace", "OrderQuery").Infof("Order query result: %+v", resp)

	// Response:
	// {message: "", {field: status, code: fail} }
	// {message: "", {field: result, code: "ORDERNOTEXIST" | "SYSTEMERROR"} }
	if r := router.client.ValidateResponse(resp); r != nil {
		if r.Field == "result" && r.Code == "ORDERNOTEXIST" {
			view.Render(w, view.NewNotFound())
			return
		}

		view.Render(w, view.NewUnprocessable(r))
		return
	}

	orderQuery := wechat.NewOrderQueryResp(resp)

	go router.model.SaveWxQueryResp(orderQuery)

	view.Render(w, view.NewResponse().SetBody(orderQuery.ToResult()))
}
