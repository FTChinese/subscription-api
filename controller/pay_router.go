package controller

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/objcoding/wxpay"
	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

const (
	apiBaseURL     = "http://www.ftacademy.cn/api"
	aliProductCode = "QUICK_MSECURITY_PAY"
)

// PayRouter is the base type used to handle shared payment operations.
type PayRouter struct {
	sandbox bool
	model   model.Env
	postman postoffice.Postman
}

// Returns notification URL for Alipay based on whether the api is run for sandbox.
func (router PayRouter) aliCallbackURL() string {
	if router.sandbox {
		return apiBaseURL + "/sandbox/callback/alipay"
	}

	return apiBaseURL + "/v1/callback/alipay"
}

func (router PayRouter) aliReturnURL() string {
	if router.sandbox {
		return apiBaseURL + "/sandbox/redirect/alipay/next-user"
	}

	return apiBaseURL + "/v1/redirect/alipay/next-user"
}

func (router PayRouter) wxCallbackURL() string {
	if router.sandbox {
		return apiBaseURL + "/sandbox/callback/wxpay"
	}

	return apiBaseURL + "/v1/callback/wxpay"
}

// AliAppPayParam builds parameters for ali app pay based on current subscription order.
func (router PayRouter) aliAppPayParam(title string, s paywall.Subscription) alipay.AliPayTradeAppPay {
	p := alipay.AliPayTradeAppPay{}
	p.NotifyURL = router.aliCallbackURL()
	p.Subject = title
	p.OutTradeNo = s.OrderID
	p.TotalAmount = s.AliNetPrice()
	p.ProductCode = aliProductCode
	p.GoodsType = "0"

	return p
}

func (router PayRouter) aliWebPayParam(title string, s paywall.Subscription) alipay.AliPayTradePagePay {
	p := alipay.AliPayTradePagePay{}
	p.NotifyURL = router.aliCallbackURL()
	p.ReturnURL = router.aliReturnURL()
	p.Subject = title
	p.OutTradeNo = s.OrderID
	p.TotalAmount = s.AliNetPrice()
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	p.GoodsType = "0"

	return p
}

// WxUniOrderParam build the parameters to request for prepay id.
func (router PayRouter) wxUniOrderParam(title, ip string, s paywall.Subscription) wxpay.Params {
	p := make(wxpay.Params)
	p.SetString("body", title)
	p.SetString("out_trade_no", s.OrderID)
	p.SetInt64("total_fee", s.WxNetPrice())
	p.SetString("spbill_create_ip", ip)
	p.SetString("notify_url", router.wxCallbackURL())
	// APP for native app
	// NATIVE for web site
	// JSAPI for web page opend inside wechat browser
	p.SetString("trade_type", "APP")

	return p
}

// SendConfirmationLetter sends a confirmation email if user logged in with FTC account.
func (router PayRouter) sendConfirmationEmail(subs paywall.Subscription) error {
	// If the FTCUserID field is null, it indicates this user
	// does not have an FTC account bound. You cannot find out
	// its email address.
	if !subs.FTCUserID.Valid {
		return nil
	}
	// Find this user's personal data
	user, err := router.model.FindUser(subs.CompoundID)

	if err != nil {
		return err
	}

	parcel, err := user.ConfirmationParcel(subs)
	if err != nil {
		return err
	}

	logger.WithField("trace", "SendConirmationLetter").Info("Send subscription confirmation letter")

	err = router.postman.Deliver(parcel)
	if err != nil {
		logger.WithField("trace", "SendConfirmationLetter").Error(err)
		return err
	}
	return nil
}
