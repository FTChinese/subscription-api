package wechat

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/view"
	"github.com/objcoding/wxpay"
)

// PayClients put various weechat payment app
// in one place.
type PayClients struct {
	clients         []Client
	indexByPlatform map[TradeType]int
	indexByID       map[string]int
}

// InitPayClients reads wechat app configuration
// from viper config.
func InitPayClients() PayClients {
	c := PayClients{
		clients:         make([]Client, 0),
		indexByPlatform: make(map[TradeType]int),
		indexByID:       make(map[string]int),
	}

	for i, cfg := range appCfgs {
		app := MustNewPayApp(cfg.Key)
		c.clients = append(c.clients, NewClient(app))

		c.indexByPlatform[cfg.Platform] = i
		// Desktop and mobile browser use the same app.
		if cfg.Platform == TradeTypeDesktop {
			c.indexByPlatform[TradeTypeMobile] = i
		}

		c.indexByID[app.AppID] = i
	}

	return c
}

// GetClientByPlatform tries to find the client used for a certain trade type.
// This is used when use is creating an order.
func (c PayClients) GetClientByPlatform(t TradeType) (Client, error) {
	i, ok := c.indexByPlatform[t]
	if !ok {
		return Client{}, fmt.Errorf("wxpay client for trade type %s not found", t)
	}

	return c.clients[i], nil
}

// GetClientByAppID searches a wechat pay app by id.
// This is used by webhook.
func (c PayClients) GetClientByAppID(id string) (Client, error) {
	i, ok := c.indexByID[id]

	if !ok {
		return Client{}, fmt.Errorf("wxpay client for app id %s not found", id)
	}

	return c.clients[i], nil
}

// Client extends wxpay.Client
type Client struct {
	*wxpay.Client
	app PayApp
}

// NewClient creats a new instance of Client.
func NewClient(app PayApp) Client {
	// Pay attention to the last parameter.
	// It should always be false because Weixin's sandbox address does not work!
	account := wxpay.NewAccount(app.AppID, app.MchID, app.APIKey, false)
	c := wxpay.NewClient(account)

	return Client{
		c,
		app,
	}
}

func (c Client) GetApp() PayApp {
	return c.app
}

func (c Client) InWxBrowserParams(u UnifiedOrderResp) InWxBrowserParams {
	p := InWxBrowserParams{
		Timestamp: GenerateTimestamp(),
		Nonce:     GenerateNonce(),
		Package:   "prepay_id=" + u.PrepayID.String,
		SignType:  "MD5",
	}

	p.Signature = c.Sign(p.ToMap(c.app.AppID))

	return p
}

// AppParams build the parameters required by native app pay.
func (c Client) AppParams(u UnifiedOrderResp) AppOrderParams {
	p := AppOrderParams{
		AppID:     c.app.AppID,
		PartnerID: u.MID.String,
		PrepayID:  u.PrepayID.String,
		Timestamp: GenerateTimestamp(),
		Nonce:     GenerateNonce(),
		Package:   "Sign=WXPay",
	}
	p.Signature = c.Sign(p.ToMap())

	return p
}

// ValidateResponse checks if wechat response is generated for our app.
// NOTE: this sdk treat return_code == FAIL as valid.
// Possible return_msg:
// appid不存在;
// 商户号mch_id与appid不匹配;
// invalid spbill_create_ip;
// spbill_create_ip参数长度有误; (Wx does not accept IPv6 like 9b5b:2ef9:6c9f:cf5:130e:984d:8958:75f9 :-<)
// Deprecate
func (c Client) ValidateResponse(params wxpay.Params) *view.Reason {

	if params.GetString("return_code") == wxpay.Fail {
		statusMsg := params.GetString("return_msg")

		reason := &view.Reason{
			Field: "status",
			Code:  "fail",
		}
		reason.SetMessage(statusMsg)

		return reason
	}

	if params.GetString("result_code") == wxpay.Fail {
		errCode := params.GetString("err_code")
		errDesc := params.GetString("err_code_des")

		reason := &view.Reason{
			Field: "result",
			Code:  errCode,
		}
		reason.SetMessage(errDesc)

		return reason
	}

	if !params.ContainsKey("appid") || (params.GetString("appid") != c.app.AppID) {

		reason := &view.Reason{
			Field: "app_id",
			Code:  view.CodeInvalid,
		}
		reason.SetMessage("Missing or wrong app id")

		return reason
	}

	if !params.ContainsKey("mch_id") || (params.GetString("mch_id") != c.app.MchID) {

		reason := &view.Reason{
			Field: "mch_id",
			Code:  view.CodeInvalid,
		}
		reason.SetMessage("Missing or wrong merchant id")

		return reason
	}

	return nil
}

func (c Client) VerifyNotification(n Notification) error {
	if ve := n.Validate(c.app); ve != nil {
		return errors.New(ve.Message)
	}

	if !c.ValidSign(n.params) {
		return errors.New("invalid sign")
	}

	return nil
}
