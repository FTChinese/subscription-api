package wechat

import (
	"errors"
	"github.com/FTChinese/go-rest/view"
	"github.com/objcoding/wxpay"
)

type PayApp struct {
	AppID  string `mapstructure:"app_id"`
	MchID  string `mapstructure:"mch_id"`
	APIKey string `mapstructure:"api_key"`
}

func (a PayApp) Ensure() error {
	if a.AppID == "" || a.MchID == "" || a.APIKey == "" {
		return errors.New("wechat pay app_id, mch_id or secret cannot be empty")
	}

	return nil
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
	if r := n.Validate(c.app); r != nil {
		return errors.New(r.GetMessage())
	}

	if !c.ValidSign(n.params) {
		return errors.New("invalid sign")
	}

	return nil
}

//func (c Client) ValidateUnifiedOrder(resp UnifiedOrderResp) *view.Reason {

//if resp.StatusCode == wxpay.Fail {
//	reason := &view.Reason{
//		Field: "status",
//		Code:  "fail",
//	}
//	reason.SetMessage(resp.StatusMessage)
//
//	return reason
//}
//
//if resp.ResultCode.String == wxpay.Fail {
//	reason := &view.Reason{
//		Field: "result",
//		Code:  resp.ErrorCode.String,
//	}
//	reason.SetMessage(resp.ErrorMessage.String)
//
//	return reason
//}

//if resp.AppID.IsZero() || resp.AppID.String != c.appID {
//	reason := &view.Reason{
//		Field: "app_id",
//		Code:  view.CodeInvalid,
//	}
//	reason.SetMessage("Missing or wrong app id")
//
//	return reason
//}
//
//if resp.MID.IsZero() || resp.MID.String != c.mchID {
//	reason := &view.Reason{
//		Field: "mch_id",
//		Code:  view.CodeInvalid,
//	}
//	reason.SetMessage("Missing or wrong merchant id")
//
//	return reason
//}

//return nil
//}

// NewPrepay creates a new Prepay instance from client appID, mchID,
// prepayID and subscription id and price.
// Signature required by wechat is not calculated at this point.
//func (c Client) NewPrepay(prepayID string, subs paywall.Subscription) AppPay {
//	nonce, _ := gorest.RandomHex(10)
//	pkg := "Sign=WXPay"
//	timestamp := fmt.Sprintf("%d", time.Now().Unix())
//
//	return AppPay{
//		FtcOrderID: subs.ID,
//		Price:      subs.ListPrice,
//		ListPrice:  subs.ListPrice,
//		Amount:   subs.Amount,
//		AppID:      c.app.AppID,
//		PartnerID:  c.app.MchID,
//		PrepayID:   prepayID,
//		Package:    pkg,
//		Nonce:      nonce,
//		Timestamp:  timestamp,
//	}
//}

// ParseResponse parses and validate wechat response.
//func (c Client) ParseResponse(r io.Reader) (wxpay.Params, error) {
//	var returnCode string
//	params := DecodeXML(r)
//
//	if params.ContainsKey("return_code") {
//		returnCode = params.GetString("return_code")
//	} else {
//		return nil, errors.New("no return_code in XML")
//	}
//
//	switch returnCode {
//	case wxpay.Fail:
//		return params, nil
//
//	case wxpay.Success:
//		if c.ValidSign(params) {
//			return params, nil
//		}
//		return nil, errors.New("invalid sign value in XML")
//
//	default:
//		return nil, errors.New("return_code value is invalid in XML")
//	}
//}
