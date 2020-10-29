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
		store.clients = append(store.clients, NewWxPayClient(app))

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
	app wechat.PayApp
	sdk *wxpay.Client
}

// NewClient creats a new instance of Client.
func NewWxPayClient(app wechat.PayApp) WxPayClient {
	// Pay attention to the last parameter.
	// It should always be false because Weixin's sandbox address does not work!
	account := wxpay.NewAccount(app.AppID, app.MchID, app.APIKey, false)
	c := wxpay.NewClient(account)
	return WxPayClient{
		app: app,
		sdk: c,
	}
}

func (c WxPayClient) GetApp() wechat.PayApp {
	return c.app
}

func (c WxPayClient) CreateOrder(pi subs.PaymentIntent, cfg wechat.UnifiedOrderConfig) (wechat.UnifiedOrder, error) {
	params := pi.WxPayParam(cfg)
	resp, err := c.sdk.UnifiedOrder(params)
	if err != nil {
		return wechat.UnifiedOrder{}, err
	}

	uo := wechat.NewUnifiedOrderResp(pi.Order.ID, resp)

	ve := uo.Validate(c.app)
	if ve != nil {
		uo.Invalid = ve
	}

	return uo, nil
}

//https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
func (c WxPayClient) QueryOrder(order subs.Order) (wechat.OrderQueryResp, error) {
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

	// If there are any errors when querying order.
	if err != nil {
		return wechat.OrderQueryResp{}, render.NewInternalError(err.Error())
	}

	return wechat.NewOrderQueryResp(respParams), nil
}

func (c WxPayClient) InWxBrowserParams(u wechat.UnifiedOrder) wechat.InWxBrowserParams {
	p := wechat.InWxBrowserParams{
		Timestamp: wechat.GenerateTimestamp(),
		Nonce:     wechat.GenerateNonce(),
		Package:   "prepay_id=" + u.PrepayID.String,
		SignType:  "MD5",
	}

	p.Signature = c.sdk.Sign(p.ToMap(c.app.AppID))

	return p
}

// AppParams build the parameters required by native app pay.
func (c WxPayClient) AppParams(u wechat.UnifiedOrder) wechat.AppOrderParams {
	p := wechat.AppOrderParams{
		AppID:     c.app.AppID,
		PartnerID: u.MID.String,
		PrepayID:  u.PrepayID.String,
		Timestamp: wechat.GenerateTimestamp(),
		Nonce:     wechat.GenerateNonce(),
		Package:   "Sign=WXPay",
	}
	p.Signature = c.sdk.Sign(p.ToMap())

	return p
}

func (c WxPayClient) ValidateNotification(n wechat.Notification) error {
	if err := n.IsStatusValid(); err != nil {
		return err
	}

	if ve := n.Validate(c.app); ve != nil {
		return errors.New(ve.Message)
	}

	if !c.sdk.ValidSign(n.RawParams) {
		return errors.New("invalid sign")
	}

	return nil
}
