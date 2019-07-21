package controller

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/sirupsen/logrus"
	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/ali"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
	"net/http"
)

const (
	apiBaseURL = "http://www.ftacademy.cn/api"
)

// PayRouter is the base type used to handle shared payment operations.
type PayRouter struct {
	model   model.Env
	postman postoffice.Postman
}

func (router PayRouter) findPlan(req *http.Request) (paywall.Plan, error) {
	t, err := GetURLParam(req, "tier").ToString()
	if err != nil {
		return paywall.Plan{}, err
	}

	c, err := GetURLParam(req, "cycle").ToString()
	if err != nil {
		return paywall.Plan{}, err
	}

	return router.model.GetCurrentPlans().FindPlan(t + "_" + c)
}

func (router PayRouter) handleOrderErr(w http.ResponseWriter, err error) {
	switch err {
	case util.ErrBeyondRenewal:
		view.Render(w, view.NewForbidden(err.Error()))

	case util.ErrDowngrade:
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
	if router.model.Sandbox {
		return apiBaseURL + "/sandbox/callback/alipay"
	}

	return apiBaseURL + "/v1/callback/alipay"
}

func (router PayRouter) wxCallbackURL() string {
	if router.model.Sandbox {
		return apiBaseURL + "/sandbox/callback/wxpay"
	}

	return apiBaseURL + "/v1/callback/wxpay"
}

// AliAppPayParam builds parameters for ali app pay based on current subscription order.
func (router PayRouter) aliAppPayParam(title string, s paywall.Subscription) alipay.AliPayTradeAppPay {
	p := alipay.AliPayTradeAppPay{}
	p.NotifyURL = router.aliCallbackURL()
	p.Subject = title
	p.OutTradeNo = s.ID
	p.TotalAmount = s.AliPrice()
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
//	p.OutTradeNo = s.ID
//	p.TotalAmount = s.AliPrice()
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
//	p.OutTradeNo = s.ID
//	p.TotalAmount = s.AliPrice()
//	p.ProductCode = ali.ProductCodeWeb.String()
//	p.GoodsType = "0"
//
//	return p
//}

// SendConfirmationLetter sends a confirmation email if user logged in with FTC account.
func (router PayRouter) sendConfirmationEmail(subs paywall.Subscription) error {
	logger := logrus.WithFields(logrus.Fields{
		"trace": "PayRouter.sendConfirmationEmail",
	})

	// If the FtcID field is null, it indicates this user
	// does not have an FTC account bound. You cannot find out
	// its email address.
	if !subs.FtcID.Valid {
		return nil
	}
	// Find this user's personal data
	user, err := router.model.FindFtcUser(subs.CompoundID)

	if err != nil {
		return err
	}

	parcel, err := user.ConfirmationParcel(subs)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("Send subscription confirmation letter")

	err = router.postman.Deliver(parcel)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}
