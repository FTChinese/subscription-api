package wechat

import (
	"errors"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"go.uber.org/zap"
)

// WxPayClient wraps wxpay.Client
type WxPayClient struct {
	app    PayApp
	sdk    *wxpay.Client
	logger *zap.Logger
}

// NewWxPayClient creates a new instance of Client.
func NewWxPayClient(app PayApp, logger *zap.Logger) WxPayClient {
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

func (c WxPayClient) GetApp() PayApp {
	return c.app
}

func (c WxPayClient) Sign(p wxpay.Params) string {
	return c.sdk.Sign(p)
}

func (c WxPayClient) VerifySignature(payload wxpay.Params) error {
	if !c.sdk.ValidSign(payload) {
		return errors.New("wxpay webhook payload: signature cannot be verified")
	}

	return nil
}

// CreateOrder at
// https://pay.weixin.qq.com/wiki/doc/api/jsapi.php?chapter=9_1
func (c WxPayClient) CreateOrder(o UnifiedOrderReq) (wxpay.Params, error) {
	return c.sdk.UnifiedOrder(o.Marshal())
}

// QueryOrder at
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
func (c WxPayClient) QueryOrder(params OrderQueryParams) (wxpay.Params, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	// Send query to Wechat server
	// Returns the parsed response as a map.
	// It checks if the response contains `return_code` key.
	// return_code == FAIL/SUCCESS only determines
	// whether the response body signature is verified.
	// Example:
	// appid:
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
	payload, err := c.sdk.OrderQuery(params.Marshal())

	sugar.Infof("wxpay query order payload: %v", payload)

	// If there are any errors when querying order.
	if err != nil {
		return nil, err
	}

	// Validate if response is correct. This does not verify the payment is successful.
	// We have to send the payment status as is to client.
	// field: return_code, code: invalid
	// field: result_code, code: invalid
	// field: app_id, code: invalid
	// field: mch_id, code: invalid
	err = c.GetApp().ValidateOrderPayload(payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (c WxPayClient) SignJSApiParams(or OrderResult) JSApiParams {
	p := NewJSApiParams(or)
	p.Signature = c.sdk.Sign(p.ToMap())

	return p
}

// SignAppParams re-sign wxpay's order to build the parameters used by native app sdk to call wechat service.
func (c WxPayClient) SignAppParams(or OrderResult) NativeAppParams {
	p := NewNativeAppParams(or)
	p.Signature = c.sdk.Sign(p.Marshal())

	return p
}

func (c WxPayClient) SDKParams(orderResp OrderResult, platform TradeType) (SDKParams, error) {
	switch platform {
	case TradeTypeDesktop:
		return SDKParams{
			DesktopQr:      null.NewString(orderResp.QRCode, orderResp.QRCode != ""),
			MobileRedirect: null.String{},
			JsApi:          JSApiParamsJSON{},
			AppSDK:         NativeAppParamsJSON{},
		}, nil

	case TradeTypeMobile:
		return SDKParams{
			DesktopQr:      null.String{},
			MobileRedirect: null.NewString(orderResp.MWebURL, orderResp.MWebURL != ""),
			JsApi:          JSApiParamsJSON{},
			AppSDK:         NativeAppParamsJSON{},
		}, nil

	case TradeTypeJSAPI:
		return SDKParams{
			DesktopQr:      null.String{},
			MobileRedirect: null.String{},
			JsApi: JSApiParamsJSON{
				JSApiParams: c.SignJSApiParams(orderResp),
			},
			AppSDK: NativeAppParamsJSON{},
		}, nil

	case TradeTypeApp:
		return SDKParams{
			DesktopQr:      null.String{},
			MobileRedirect: null.String{},
			JsApi:          JSApiParamsJSON{},
			AppSDK: NativeAppParamsJSON{
				NativeAppParams: c.SignAppParams(orderResp),
			},
		}, nil

	default:
		return SDKParams{}, errors.New("unknown wechat pay platform")
	}
}
