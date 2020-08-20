package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/models/plan"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/repository/products"
	"github.com/FTChinese/subscription-api/repository/readerrepo"
	"github.com/FTChinese/subscription-api/repository/subrepo"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"net/http"
)

// PayRouter is the base type used to handle shared payment operations.
type PayRouter struct {
	subEnv    subrepo.SubEnv
	readerEnv readerrepo.ReaderEnv
	prodRepo  products.Env
	postman   postoffice.PostOffice
	config    config.BuildConfig
}

func NewBasePayRouter(db *sqlx.DB, c *cache.Cache, b config.BuildConfig, p postoffice.PostOffice) PayRouter {
	return PayRouter{
		subEnv:    subrepo.NewSubEnv(db, c, b),
		readerEnv: readerrepo.NewReaderEnv(db, b),
		prodRepo:  products.NewEnv(db, c),
		postman:   p,
		config:    b,
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

// Centralized error handling after order creation.
// It handles the errors propagated from Membership.SubsKind(),
func (router PayRouter) handleOrderErr(w http.ResponseWriter, err error) {
	switch err {
	// When the order is used to renew but not allowed.
	case subs.ErrRenewalForbidden:
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: err.Error(),
			Field:   "renewal",
			Code:    render.CodeInvalid,
		})

	case subs.ErrDowngradeForbidden:
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: err.Error(),
			Field:   "downgrade",
			Code:    render.CodeInvalid,
		})

	default:
		_ = render.New(w).DBError(err)
	}
}

// SendConfirmationLetter sends a confirmation email if user logged in with FTC account.
func (router PayRouter) sendConfirmationEmail(order subs.Order) error {
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
	switch order.Kind {
	case enum.OrderKindCreate:
		parcel, err = letter.NewSubParcel(account, order)

	case enum.OrderKindRenew:
		parcel, err = letter.NewRenewalParcel(account, order)

	case enum.OrderKindUpgrade:
		up, err := router.readerEnv.LoadUpgradeSchema(order.ID)
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

func (router PayRouter) sendFreeUpgradeEmail(order subs.Order, wallet subs.Wallet) error {
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
