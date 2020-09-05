package controller

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/repository/products"
	"github.com/FTChinese/subscription-api/repository/readerrepo"
	"github.com/FTChinese/subscription-api/repository/subrepo"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"net/http"
)

// PayRouter is the base type used to handle shared payment operations.
type PayRouter struct {
	subEnv    subrepo.Env
	readerEnv readerrepo.Env
	prodRepo  products.Env
	postman   postoffice.PostOffice
	config    config.BuildConfig
	logger    *zap.Logger
}

func NewBasePayRouter(db *sqlx.DB, c *cache.Cache, b config.BuildConfig, p postoffice.PostOffice) PayRouter {
	l, _ := zap.NewProduction()

	return PayRouter{
		subEnv:    subrepo.NewEnv(db, c, b),
		readerEnv: readerrepo.NewEnv(db, b),
		prodRepo:  products.NewEnv(db, c),
		postman:   p,
		config:    b,
		logger:    l,
	}
}

// Only when the user has ftc account, and
// query parameter has `test=true` will
// we search db to see whether it is actually
// a test account.
func (router PayRouter) isTestAccount(ids reader.MemberID, req *http.Request) bool {
	isTest := ids.FtcID.Valid && req.FormValue("test") == "true"

	if !isTest {
		return false
	}

	found, err := router.readerEnv.SandboxUserExists(ids.FtcID.String)
	if err != nil {
		return false
	}

	return found
}

// Centralized error handling after order creation.
// It handles the errors propagated from Membership.AliWxSubsKind(),
func (router PayRouter) handleOrderErr(w http.ResponseWriter, err error) {
	var ve *render.ValidationError
	if errors.As(err, &ve) {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	_ = render.New(w).DBError(err)
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
	account, err := router.readerEnv.AccountByFtcID(order.FtcID.String)

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
		pos, err := router.subEnv.ListProratedOrders(order.ID)
		if err != nil {
			return err
		}
		parcel, err = letter.NewUpgradeParcel(account, order, pos)
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
