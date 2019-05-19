package controller

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
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

func (router PayRouter) subsKind(user paywall.User, p paywall.Plan) (paywall.SubsKind, error) {
	member, err := router.model.RetrieveMember(user)

	if err != nil {
		// User is not a member yet.
		if err == sql.ErrNoRows {
			return paywall.SubsKindCreate, nil
		}

		return paywall.SubsKindDeny, err
	}

	// User is a member but expired.
	if member.IsExpired() {
		return paywall.SubsKindCreate, nil
	}

	if member.Tier == enum.TierStandard && p.Tier == enum.TierPremium {
		return paywall.SubsKindUpgrade, nil
	}

	if member.IsRenewAllowed() {
		return paywall.SubsKindRenew, nil
	}

	return paywall.SubsKindDeny, errRenewalForbidden
}

func (router PayRouter) createOrder(user paywall.User, p paywall.Plan) (paywall.Subscription, error) {
	subs, err := paywall.NewSubs(user, p)
	if err != nil {
		return subs, err
	}

	subsKind, err := router.subsKind(user, p)
	if err != nil {
		return subs, err
	}

	if subsKind == paywall.SubsKindUpgrade {
		orders, err := router.model.FindProration(user)
		if err != nil {
			return subs, err
		}

		up := paywall.NewUpgradePlan(p).
			SetProration(orders).
			CalculatePayable()

		subs = subs.WithUpgrade(up)
	} else {
		subs.Kind = subsKind
	}

	return subs, nil
}

// Returns notification URL for Alipay based on whether the api is run for sandbox.
func (router PayRouter) aliCallbackURL() string {
	if router.sandbox {
		return apiBaseURL + "/sandbox/callback/alipay"
	}

	return apiBaseURL + "/v1/callback/alipay"
}

func (router PayRouter) aliReturnURL(override string) string {
	if override != "" {
		return override
	}

	if router.sandbox {
		return apiBaseURL + "/sandbox/redirect/alipay/done"
	}

	return apiBaseURL + "/v1/redirect/alipay/done"
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
