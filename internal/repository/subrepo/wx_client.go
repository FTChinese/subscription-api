package subrepo

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/objcoding/wxpay"
	"go.uber.org/zap"
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
		return WxPayClient{}, fmt.Errorf("wxpay client for trade type %s not found", t)
	}

	return s.clients[i], nil
}

// GetClientByAppID searches a wechat pay app by id.
// This is used by webhook.
func (s WxPayClientStore) ClientByAppID(id string) (WxPayClient, error) {
	i, ok := s.indexByID[id]

	if !ok {
		return WxPayClient{}, fmt.Errorf("wxpay client for app id %s not found", id)
	}

	return s.clients[i], nil
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

//https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
func (c WxPayClient) QueryOrder(order subs.Order) (wechat.OrderQueryResp, error) {
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
	// appid:***REMOVED***
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
	respParams, err := c.sdk.OrderQuery(reqParams)

	sugar.Infof("Query wx order: %v", respParams)

	// If there are any errors when querying order.
	if err != nil {
		return wechat.OrderQueryResp{}, render.NewInternalError(err.Error())
	}

	return wechat.NewOrderQueryResp(respParams), nil
}

func (c WxPayClient) ValidateWebhook(payload wechat.Notification) error {

	err := payload.Validate(c.app)
	if err != nil {
		return err
	}

	if !c.sdk.ValidSign(payload.RawParams) {
		return errors.New("invalid sign")
	}

	return nil
}

func (c WxPayClient) SignJSApiParams(or wechat.OrderResp) wechat.JSApiParams {
	p := wechat.NewJSApiParams(or)
	p.Signature = c.sdk.Sign(p.ToMap())

	return p
}

// SignAppParams build the parameters required by native app pay.
func (c WxPayClient) SignAppParams(or wechat.OrderResp) wechat.NativeAppParams {
	p := wechat.NewNativeAppParams(or)
	p.Signature = c.sdk.Sign(p.ToMap())

	return p
}
