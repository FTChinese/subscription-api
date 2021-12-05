package subrepo

import (
	"errors"
	"fmt"
	subs2 "github.com/FTChinese/subscription-api/internal/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/objcoding/wxpay"
	"go.uber.org/zap"
	"net/http"
)

// PayClients put various weechat payment app
// in one place.
type WxPayClientStore struct {
	clients         []WxPayClient
	indexByPlatform map[wechat.TradeType]int
	indexByID       map[string]int
	logger          *zap.Logger
}

func NewWxClientStore(apps []wechat.PayApp, logger *zap.Logger) WxPayClientStore {
	store := WxPayClientStore{
		clients:         make([]WxPayClient, 0),
		indexByPlatform: make(map[wechat.TradeType]int),
		indexByID:       make(map[string]int),
		logger:          logger,
	}

	for i, app := range apps {
		store.clients = append(store.clients, NewWxPayClient(app, logger))

		store.indexByPlatform[app.Platform] = i
		// Desktop and mobile browser use the same app.
		if app.Platform == wechat.TradeTypeDesktop {
			store.indexByPlatform[wechat.TradeTypeMobile] = i
		}

		store.indexByID[app.AppID] = i
	}

	return store
}

// GetClientByPlatform tries to find the client used for a certain trade type.
// This is used when use is creating an order.
func (s WxPayClientStore) ClientByPlatform(t wechat.TradeType) (WxPayClient, error) {
	i, ok := s.indexByPlatform[t]
	if !ok {
		return WxPayClient{}, fmt.Errorf("wxpay client: cannot find app for trade type %s", t)
	}

	return s.clients[i], nil
}

// GetClientByAppID searches a wechat pay app by id.
// This is used by webhook.
func (s WxPayClientStore) ClientByAppID(id string) (WxPayClient, error) {
	i, ok := s.indexByID[id]

	if !ok {
		return WxPayClient{}, fmt.Errorf("wxpay client: cannot find app %s", id)
	}

	return s.clients[i], nil
}

func (s WxPayClientStore) GetWebhookPayload(req *http.Request) (wechat.Notification, error) {
	defer s.logger.Sync()
	sugar := s.logger.Sugar()

	defer req.Body.Close()

	raw, err := wechat.DecodeXML(req.Body)
	if err != nil {
		sugar.Error()
		return wechat.Notification{}, err
	}

	sugar.Infof("wxpay webhook raw payload: %+v", raw)

	payload := wechat.NewNotification(raw)

	if payload.IsBadRequest() {
		return wechat.Notification{}, fmt.Errorf("wxpay webhook payload: %s", payload.BadRequestMsg())
	}

	if payload.AppID.IsZero() {
		return wechat.Notification{}, fmt.Errorf("wxpay webhook payload: missing appid")
	}

	client, err := s.ClientByAppID(payload.AppID.String)
	if err != nil {
		return wechat.Notification{}, err
	}

	if !client.sdk.ValidSign(raw) {
		return wechat.Notification{}, errors.New("wxpay webhook payload: signature cannot be verified")
	}

	return payload, nil
}

func (s WxPayClientStore) QueryOrderRaw(order subs2.Order) (wxpay.Params, error) {
	client, err := s.ClientByAppID(order.WxAppID.String)
	if err != nil {
		return nil, err
	}

	return client.queryOrderRaw(order.ID)
}

func (s WxPayClientStore) VerifyPayment(order subs2.Order) (subs2.PaymentResult, error) {
	defer s.logger.Sync()
	sugar := s.logger.Sugar()

	client, err := s.ClientByAppID(order.WxAppID.String)
	if err != nil {
		sugar.Error(err)
		return subs2.PaymentResult{}, err
	}

	return client.VerifyPayment(order)
}

// Client extends wxpay.Client
type WxPayClient struct {
	app    wechat.PayApp
	sdk    *wxpay.Client
	logger *zap.Logger
}

// NewClient creats a new instance of Client.
func NewWxPayClient(app wechat.PayApp, logger *zap.Logger) WxPayClient {
	// Pay attention to the last parameter.
	// It should always be false because Weixin's sandbox address does not work!
	account := wxpay.NewAccount(app.AppID, app.MchID, app.APIKey, false)
	c := wxpay.NewClient(account)
	return WxPayClient{
		app:    app,
		sdk:    c,
		logger: logger,
	}
}

func (c WxPayClient) GetApp() wechat.PayApp {
	return c.app
}

func (c WxPayClient) Sign(p wxpay.Params) string {
	return c.sdk.Sign(p)
}

func (c WxPayClient) CreateOrder(o wechat.OrderReq) (wechat.OrderResp, error) {

	resp, err := c.sdk.UnifiedOrder(o.ToMap())
	if err != nil {
		return wechat.OrderResp{}, err
	}

	return wechat.NewOrderResp(o.SellerOrderID, resp), nil
}

func (c WxPayClient) queryOrderRaw(id string) (wxpay.Params, error) {
	reqParams := make(wxpay.Params)
	reqParams.SetString("out_trade_no", id)

	return c.sdk.OrderQuery(reqParams)
}

//https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
func (c WxPayClient) QueryOrder(order subs2.Order) (wechat.OrderQueryResp, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	reqParams := make(wxpay.Params)
	reqParams.SetString("out_trade_no", order.ID)

	// Send query to Wechat server
	// Returns the parsed response as a map.
	// It checks if the response contains `return_code` key.
	// return_code == FAIL/SUCCESS only determines
	// whether the response body signature is verified.
	// Example:
	// appid:wxacddf1c20516eb69
	// device_info: mch_id:1504993271
	// nonce_str:9dmEFWFU5ooB9dMN
	// out_trade_no:FT9F67C5CC9F47CF65
	// result_code:SUCCESS
	// return_code:SUCCESS
	// return_msg:OK
	// sign:538529EAEE06FE61ECE379C699437B37
	// total_fee:25800
	// trade_state:NOTPAY
	// trade_state_desc:订单未支付
	raw, err := c.sdk.OrderQuery(reqParams)

	sugar.Infof("wxpay client raw query order: %v", raw)

	// If there are any errors when querying order.
	if err != nil {
		return wechat.OrderQueryResp{}, err
	}

	return wechat.NewOrderQueryResp(raw), nil
}

func (c WxPayClient) VerifyPayment(order subs2.Order) (subs2.PaymentResult, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	ordResp, err := c.QueryOrder(order)
	if err != nil {
		sugar.Error(err)
		return subs2.PaymentResult{}, err
	}

	// Validate if response is correct. This does not verify the payment is successful.
	// We have to send the payment status as is to client.
	// field: return_code, code: invalid
	// field: result_code, code: invalid
	// field: app_id, code: invalid
	// field: mch_id, code: invalid
	err = ordResp.Validate(c.app)
	if err != nil {
		sugar.Error(err)
		return subs2.PaymentResult{}, err
	}

	return subs2.NewWxPayResult(ordResp), nil
}

func (c WxPayClient) SignJSApiParams(or wechat.OrderResp) wechat.JSApiParams {
	p := wechat.NewJSApiParams(or)
	p.Signature = c.sdk.Sign(p.ToMap())

	return p
}

// SignAppParams re-sign wxpay's order to build the parameters used by native app sdk to call wechat service.
func (c WxPayClient) SignAppParams(or wechat.OrderResp) wechat.NativeAppParams {
	p := wechat.NewNativeAppParams(or)
	p.Signature = c.sdk.Sign(p.ToMap())

	return p
}
