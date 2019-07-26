package controller

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/sirupsen/logrus"
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
		return apiBaseURL + "/sandbox/webhook/alipay"
	}

	return apiBaseURL + "/v1/webhook/alipay"
}

func (router PayRouter) wxCallbackURL() string {
	if router.model.Sandbox {
		return apiBaseURL + "/sandbox/webhook/wxpay"
	}

	return apiBaseURL + "/v1/webhook/wxpay"
}

// SendConfirmationLetter sends a confirmation email if user logged in with FTC account.
func (router PayRouter) sendConfirmationEmail(subs paywall.Subscription) error {
	logger := logrus.WithFields(logrus.Fields{
		"trace": "PayRouter.sendConfirmationEmail",
	})

	// If the FtcID field is null, it indicates this user
	// does not have an FTC account bound. You cannot find out
	// its email address.
	if !subs.User.FtcID.Valid {
		return nil
	}
	// Find this user's personal data
	ftcUser, err := router.model.FindFtcUser(subs.User.CompoundID)

	if err != nil {
		return err
	}

	var parcel postoffice.Parcel
	switch subs.Usage {
	case paywall.SubsKindCreate:
		parcel, err = ftcUser.NewSubParcel(subs)

	case paywall.SubsKindRenew:
		parcel, err = ftcUser.RenewSubParcel(subs)

	case paywall.SubsKindUpgrade:
		up, err := router.model.LoadUpgradeSource(subs.ID)
		if err != nil {
			return err
		}
		parcel, err = ftcUser.UpgradeSubParcel(subs, up)
	}

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
