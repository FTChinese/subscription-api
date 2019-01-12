package wechat

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/util"

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
