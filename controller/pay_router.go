package controller

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/letter"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository/rederrepo"
	"gitlab.com/ftchinese/subscription-api/repository/subrepo"
	"net/http"
)

// PayRouter is the base type used to handle shared payment operations.
type PayRouter struct {
	subEnv    subrepo.SubEnv
	readerEnv rederrepo.ReaderEnv
	postman   postoffice.Postman
}

func NewBasePayRouter(subEnv subrepo.SubEnv, readerEnv rederrepo.ReaderEnv, p postoffice.Postman) PayRouter {
	return PayRouter{
		subEnv:    subEnv,
		readerEnv: readerEnv,
		postman:   p,
	}
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

	return router.subEnv.GetCurrentPlans().FindPlan(t + "_" + c)
}

func (router PayRouter) handleOrderErr(w http.ResponseWriter, err error) {
	switch err {
	case util.ErrBeyondRenewal:
		_ = view.Render(w, view.NewForbidden(err.Error()))

	case util.ErrDowngrade:
		r := view.NewReason()
		r.Field = "downgrade"
		r.Code = view.CodeInvalid
		r.SetMessage(err.Error())
		_ = view.Render(w, view.NewUnprocessable(r))

	default:
		_ = view.Render(w, view.NewDBFailure(err))
	}
}

// SendConfirmationLetter sends a confirmation email if user logged in with FTC account.
func (router PayRouter) sendConfirmationEmail(order subscription.Order) error {
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
	account, err := router.readerEnv.FindAccountByFtcID(order.FtcID.String)

	if err != nil {
		return err
	}

	var parcel postoffice.Parcel
	switch order.Usage {
	case plan.SubsKindCreate:
		parcel, err = letter.NewSubParcel(account, order)

	case plan.SubsKindRenew:
		parcel, err = letter.NewRenewalParcel(account, order)

	case plan.SubsKindUpgrade:
		up, err := router.readerEnv.LoadUpgradeSchema(order.UpgradeSchemaID.String)
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
