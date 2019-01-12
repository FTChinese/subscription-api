package wechat

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/view"

	log "github.com/sirupsen/logrus"
)

var logger = log.WithField("package", "wechat")

// Client extends wxpay.Client
type Client struct {
	appID  string
	mchID  string
	apiKey string
	*wxpay.Client
}

// NewClient creats a new instance of Client.
func NewClient(appID, mchID, apiKey string) Client {
	account := wxpay.NewAccount(appID, mchID, apiKey, false)
	c := wxpay.NewClient(account)

	return Client{
		appID,
		mchID,
		apiKey,
		c,
	}
}

// ParseResponse parses and validate wechat response.
func (c Client) ParseResponse(r io.Reader) (wxpay.Params, error) {
	var returnCode string
	params := DecodeXML(r)

	if params.ContainsKey("return_code") {
		returnCode = params.GetString("return_code")
	} else {
		return nil, errors.New("no return_code in XML")
	}

	switch returnCode {
	case wxpay.Fail:
		return nil, errors.New("wx notification failed")

	case wxpay.Success:
		if c.ValidSign(params) {
			return params, nil
		}
		return nil, errors.New("invalid sign value in XML")

	default:
		return nil, errors.New("return_code value is invalid in XML")
	}
}

// VerifyIdentity checks if wechat response is generated for our app.
func (c Client) VerifyIdentity(params wxpay.Params) bool {
	if !params.ContainsKey("appid") || (params.GetString("appid") != c.appID) {
		logger.WithField("trace", "VerifyIdentity").Error("Missing or wrong appid")
		return false
	}

	if !params.ContainsKey("mch_id") || (params.GetString("mch_id") != c.mchID) {
		logger.WithField("trace", "VerifyIdentity").Error("Missing or wrong mch_id")
		return false
	}

	return true
}

// BuildPrepayOrder for client.
func (c Client) BuildPrepayOrder(orderID string, price float64, prepayID string) PrepayOrder {
	nonce, _ := util.RandomHex(10)
	pkg := "Sign=WXPay"
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	p := make(wxpay.Params)
	p["appid"] = c.appID
	p["partnerid"] = c.mchID
	p["prepayid"] = prepayID
	p["package"] = pkg
	p["noncestr"] = nonce
	p["timestamp"] = timestamp

	h := c.Sign(p)

	return PrepayOrder{
		FtcOrderID: orderID,
		Price:      price,
		AppID:      c.appID,
		PartnerID:  c.mchID,
		PrepayID:   prepayID,
		Package:    pkg,
		Nonce:      nonce,
		Timestamp:  timestamp,
		Signature:  h,
	}
}

// ValidateResponse verifies if wechat response is valid.
//
// Example response:
// return_code:SUCCESS|FAIL
// return_msg:OK
//
// Present only if return_code == SUCCESS
// appid:wx......
// mch_id:........
// nonce_str:8p8ZlUBkLsFPxC6g
// sign:DB68F0D9F193D499DF9A2EDBFFEAF312
// result_code:SUCCESS|FAIL
// err_code
// err_code_des
//
// Present only if returnd_code == SUCCESS and result_code == SUCCCESS
// trade_type:APP
// prepay_id:wx20125006086590be8d9519f40090763508
// NOTE: this sdk treat return_code == FAIL as valid.
// Possible return_msg:
// appid不存在;
// 商户号mch_id与appid不匹配;
// invalid spbill_create_ip;
// spbill_create_ip参数长度有误; (Wx does not accept IPv6 like 9b5b:2ef9:6c9f:cf5:130e:984d:8958:75f9 :-<)
func ValidateResponse(resp wxpay.Params) *view.Reason {
	if resp.GetString("return_code") == wxpay.Fail {
		returnMsg := resp.GetString("return_msg")
		logger.
			WithField("trace", "ValidateResponse").
			Errorf("return_code is FAIL. return_msg: %s", returnMsg)

		reason := &view.Reason{
			Field: "return_code",
			Code:  "fail",
		}
		reason.SetMessage(returnMsg)

		return reason
	}

	if resp.GetString("result_code") == wxpay.Fail {
		errCode := resp.GetString("err_code")
		errCodeDes := resp.GetString("err_code_des")

		logger.WithField("trace", "ValidateResponse").
			WithField("err_code", errCode).
			WithField("err_code_des", errCodeDes).
			Error("Wx unified order result failed")

		reason := &view.Reason{
			Field: "result_code",
			Code:  errCode,
		}
		reason.SetMessage(errCodeDes)

		return reason
	}

	return nil
}
