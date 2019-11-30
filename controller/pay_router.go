package controller

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/letter"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository"
	"net/http"
)

const (
	apiBaseURL = "http://www.ftacademy.cn/api"
)

// PayRouter is the base type used to handle shared payment operations.
type PayRouter struct {
	env     repository.Env
	postman postoffice.Postman
}

func (router PayRouter) findPlan(req *http.Request) (plan.Plan, error) {
	t, err := GetURLParam(req, "tier").ToString()
	if err != nil {
		return plan.Plan{}, err
	}

	c, err := GetURLParam(req, "cycle").ToString()
	if err != nil {
		return plan.Plan{}, err
	}

	return router.env.GetCurrentPlans().FindPlan(t + "_" + c)
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
	if router.env.Sandbox {
		return apiBaseURL + "/sandbox/webhook/alipay"
	}

	return apiBaseURL + "/v1/webhook/alipay"
}

func (router PayRouter) wxCallbackURL() string {
	if router.env.Sandbox {
		return apiBaseURL + "/sandbox/webhook/wxpay"
	}

	return apiBaseURL + "/v1/webhook/wxpay"
}

// SendConfirmationLetter sends a confirmation email if user logged in with FTC account.
func (router PayRouter) sendConfirmationEmail(order paywall.Order) error {
	logger := logrus.WithFields(logrus.Fields{
		"trace": "PayRouter.sendConfirmationEmail",
	})

	// If the FtcID field is null, it indicates this user
	// does not have an FTC account linked. You cannot find out
	// its email address.
	if !order.FtcID.Valid {
		return nil
	}
	// Find this user's personal data
	account, err := router.env.FindFtcUser(order.FtcID.String)

	if err != nil {
		return err
	}

	var parcel postoffice.Parcel
	switch order.Usage {
	case paywall.SubsKindCreate:
		parcel, err = letter.NewSubParcel(account, order)

	case paywall.SubsKindRenew:
		parcel, err = letter.NewRenewalParcel(account, order)

	case paywall.SubsKindUpgrade:
		up, err := router.loadUpgradePlan(order.UpgradeID.String)
		if err != nil {
			return err
		}
		parcel, err = letter.NewUpgradeParcel(account, order, up)
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

func (router PayRouter) loadUpgradePlan(upgradeID string) (paywall.UpgradePlan, error) {
	up, err := router.env.RetrieveUpgradePlan(upgradeID)
	if err != nil {
		return up, err
	}

	sources, err := router.env.RetrieveProratedOrders(upgradeID)
	if err != nil {
		return up, err
	}

	up.Data = sources

	return up, nil
}
