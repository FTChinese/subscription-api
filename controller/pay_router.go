package controller

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/ali"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"net/http"
)

const (
	apiBaseURL = "http://www.ftacademy.cn/api"
)

// PayRouter is the base type used to handle shared payment operations.
type PayRouter struct {
	sandbox bool
	model   model.Env
	postman postoffice.Postman
}

func (router PayRouter) findPlan(req *http.Request) (paywall.Plan, error) {
	tier, err := GetURLParam(req, "tier").ToString()
	if err != nil {
		return paywall.Plan{}, err
	}

	cycle, err := GetURLParam(req, "cycle").ToString()
	if err != nil {
		return paywall.Plan{}, err
	}

	return router.model.GetCurrentPricing().FindPlan(tier, cycle)
}

func (router PayRouter) handleOrderErr(w http.ResponseWriter, err error) {
	switch err {
	case paywall.ErrBeyondRenewal:
		view.Render(w, view.NewForbidden(err.Error()))

	case paywall.ErrDowngrade:
		r := view.NewReason()
		r.Field = "downgrade"
		r.Code = view.CodeInvalid
		r.SetMessage(err.Error())
		view.Render(w, view.NewUnprocessable(r))

	default:
		view.Render(w, view.NewDBFailure(err))
	}
}

// Returns notification URL for Alipay based on whether the api is run for sandbox.
func (router PayRouter) aliCallbackURL() string {
	if router.sandbox {
		return apiBaseURL + "/sandbox/callback/alipay"
	}

	return apiBaseURL + "/v1/callback/alipay"
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
	p.ProductCode = ali.ProductCodeApp.String()
	p.GoodsType = "0"

	return p
}

// The used by this one is exactly the same as `aliWapPayParam` except the return types are different.
// They are created separately because `alipay` package
// requires different data types.
//func (router PayRouter) aliDesktopPayParam(title string, s paywall.Subscription) alipay.AliPayTradePagePay {
//	p := alipay.AliPayTradePagePay{}
//	p.NotifyURL = router.aliCallbackURL()
//	p.ReturnURL = router.aliReturnURL()
//	p.Subject = title
//	p.OutTradeNo = s.OrderID
//	p.TotalAmount = s.AliNetPrice()
//	p.ProductCode = ali.ProductCodeWeb.String()
//	p.GoodsType = "0"
//
//	return p
//}

//func (router PayRouter) aliWapPayParam(title string, s paywall.Subscription) alipay.AliPayTradeWapPay {
//	p := alipay.AliPayTradeWapPay{}
//	p.NotifyURL = router.aliCallbackURL()
//	p.ReturnURL = router.aliReturnURL()
//	p.Subject = title
//	p.OutTradeNo = s.OrderID
//	p.TotalAmount = s.AliNetPrice()
//	p.ProductCode = ali.ProductCodeWeb.String()
//	p.GoodsType = "0"
//
//	return p
//}

// SendConfirmationLetter sends a confirmation email if user logged in with FTC account.
func (router PayRouter) sendConfirmationEmail(subs paywall.Subscription) error {
	// If the FTCUserID field is null, it indicates this user
	// does not have an FTC account bound. You cannot find out
	// its email address.
	if !subs.FtcID.Valid {
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
