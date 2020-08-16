package controller

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/models/letter"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/pkg/config"
	"gitlab.com/ftchinese/subscription-api/repository/readerrepo"
	"gitlab.com/ftchinese/subscription-api/repository/subrepo"
	"net/http"
)

// PayRouter is the base type used to handle shared payment operations.
type PayRouter struct {
	subEnv    subrepo.SubEnv
	readerEnv readerrepo.ReaderEnv
	postman   postoffice.PostOffice
}

func NewBasePayRouter(db *sqlx.DB, c *cache.Cache, b config.BuildConfig, p postoffice.PostOffice) PayRouter {
	return PayRouter{
		subEnv:    subrepo.NewSubEnv(db, c, b),
		readerEnv: readerrepo.NewReaderEnv(db, b),
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
	log := logger.WithField("trace", "PayRouter.sendConfirmationEmail")

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
		log.Error(err)
		return err
	}

	log.Info("Send subscription confirmation letter")

	err = router.postman.Deliver(parcel)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (router PayRouter) sendFreeUpgradeEmail(order subscription.Order, wallet subscription.Wallet) error {
	log := logger.WithField("trace", "PayRouter.sendFreeUpgradeEmail")

	// Find this user's personal data
	account, err := router.readerEnv.FindAccountByFtcID(order.FtcID.String)

	if err != nil {
		return err
	}

	parcel, err := letter.NewFreeUpgradeParcel(account, order, wallet)

	err = router.postman.Deliver(parcel)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
