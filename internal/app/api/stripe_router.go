package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/internal/repository"
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/internal/repository/stripeenv"
	"github.com/FTChinese/subscription-api/internal/stripeclient"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"net/http"
)

type StripeRouter struct {
	signingKey     string
	publishableKey string
	readerRepo     shared.ReaderCommon
	stripeRepo     stripeenv.Env
	cacheRepo      repository.CacheRepo
	logger         *zap.Logger
	live           bool
}

func NewStripeRouter(
	dbs db.ReadWriteMyDBs,
	c *cache.Cache,
	logger *zap.Logger,
	live bool,
) StripeRouter {
	return StripeRouter{
		signingKey: config.MustStripeWebhookKey().
			Pick(live),
		publishableKey: config.MustStripePubKey().
			Pick(live),
		readerRepo: shared.NewReaderCommon(dbs),
		stripeRepo: stripeenv.New(
			stripeclient.New(live, logger),
			repository.NewStripeRepo(dbs, logger),
		),
		cacheRepo: repository.NewCacheRepo(c),
		logger:    logger,
		live:      live,
	}
}

func (router StripeRouter) handleSubsResult(result stripe.SubsSuccess) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	err := router.stripeRepo.UpsertSubs(result.Subs, true)
	if err != nil {
		sugar.Error(err)
	}

	err = router.stripeRepo.UpsertInvoice(result.Subs.LatestInvoice)
	if err != nil {
		sugar.Error(err)
	}

	err = router.stripeRepo.UpsertPaymentIntent(result.Subs.PaymentIntent)
	if err != nil {
		sugar.Error(err)
	}

	if !result.Versioned.IsZero() {
		err := router.readerRepo.VersionMembership(result.Versioned)
		if err != nil {
			sugar.Error(err)
		}
	}
}

func handleSubsError(w http.ResponseWriter, err error) {
	switch err {
	case reader.ErrTrialUpgradeForbidden:
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: err.Error(),
			Field:   "trial_upgrade",
			Code:    render.CodeInvalid,
		})

	case reader.ErrAlreadyStripeSubs,
		reader.ErrAlreadyAppleSubs,
		reader.ErrAlreadyB2BSubs:
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: err.Error(),
			Field:   "membership",
			Code:    render.CodeAlreadyExists,
		})

	case reader.ErrUnknownPaymentMethod:
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: err.Error(),
			Field:   "payMethod",
			Code:    render.CodeInvalid,
		})

	default:
		_ = xhttp.HandleStripeErr(w, err)
	}
}

func (router StripeRouter) findCartItem(params stripe.SubsParams) (reader.CartItemStripe, error) {
	paywall, err := router.cacheRepo.LoadPaywall(router.live)
	if err == nil {
		item, err := params.BuildCartItem(paywall.StripePrices)
		if err == nil {
			return item, nil
		}
	}

	item, err := router.stripeRepo.LoadCheckoutItem(params)
	if err != nil {
		return reader.CartItemStripe{}, err
	}

	// Save to our database if not saved yet.
	if item.AnyFromStripe() {
		go func() {
			defer router.logger.Sync()
			sugar := router.logger.Sugar()

			if item.Recurring.IsFromStripe {
				err := router.stripeRepo.UpsertPrice(item.Recurring)
				if err != nil {
					sugar.Error(err)
				}
			}

			if item.Introductory.IsFromStripe {
				err := router.stripeRepo.UpsertPrice(item.Introductory)
				if err != nil {
					sugar.Error(err)
				}
			}
		}()
	}

	return item, nil
}
