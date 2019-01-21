package wechat

import (
	"errors"
	"fmt"
	"io"
	"time"

	"gitlab.com/ftchinese/subscription-api/paywall"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/view"
	"github.com/objcoding/wxpay"

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

// ValidateResponse checks if wechat response is generated for our app.
// NOTE: this sdk treat return_code == FAIL as valid.
// Possible return_msg:
// appid不存在;
// 商户号mch_id与appid不匹配;
// invalid spbill_create_ip;
// spbill_create_ip参数长度有误; (Wx does not accept IPv6 like 9b5b:2ef9:6c9f:cf5:130e:984d:8958:75f9 :-<)
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

	if !params.ContainsKey("appid") || (params.GetString("appid") != c.appID) {

		reason := &view.Reason{
			Field: "app_id",
			Code:  view.CodeInvalid,
		}
		reason.SetMessage("Missing or wrong app id")

		return reason
	}

	if !params.ContainsKey("mch_id") || (params.GetString("mch_id") != c.mchID) {

		reason := &view.Reason{
			Field: "mch_id",
			Code:  view.CodeInvalid,
		}
		reason.SetMessage("Missing or wrong merchant id")

		return reason
	}

	return nil
}

// NewPrepay creates a new Prepay instance from client appID, mchID,
// prepayID and subscription id and price.
// Signature required by wechat is not calculated at this point.
func (c Client) NewPrepay(prepayID string, subs paywall.Subscription) Prepay {
	nonce, _ := gorest.RandomHex(10)
	pkg := "Sign=WXPay"
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	return Prepay{
		FtcOrderID: subs.OrderID,
		Price:      subs.ListPrice,
		ListPrice:  subs.ListPrice,
		NetPrice:   subs.NetPrice,
		AppID:      c.appID,
		PartnerID:  c.mchID,
		PrepayID:   prepayID,
		Package:    pkg,
		Nonce:      nonce,
		Timestamp:  timestamp,
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
		return params, nil

	case wxpay.Success:
		if c.ValidSign(params) {
			return params, nil
		}
		return nil, errors.New("invalid sign value in XML")

	default:
		return nil, errors.New("return_code value is invalid in XML")
	}
}
