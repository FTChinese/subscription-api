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
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"net/http"
)

type PublishableKey struct {
	Key  string `json:"key"`
	Live bool   `json:"live"`
}

type StripeRoutes struct {
	signingKey     string
	publishableKey string
	readerRepo     shared.ReaderCommon
	stripeRepo     stripeenv.Env
	cacheRepo      repository.CacheRepo
	logger         *zap.Logger
	live           bool
}

func NewStripeRoutes(
	dbs db.ReadWriteMyDBs,
	c *cache.Cache,
	logger *zap.Logger,
	live bool,
) StripeRoutes {
	return StripeRoutes{
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

func (routes StripeRoutes) PublishableKey(w http.ResponseWriter, req *http.Request) {
	_ = render.New(w).OK(PublishableKey{
		Key:  routes.publishableKey,
		Live: routes.live,
	})
}

func (routes StripeRoutes) handleSubsResult(result stripe.SubsSuccess) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	err := routes.stripeRepo.UpsertSubs(result.Subs, true)
	if err != nil {
		sugar.Error(err)
	}

	if !result.Subs.Discount.IsZero() {
		err := routes.stripeRepo.UpsertDiscount(result.Subs.Discount.Discount)
		if err != nil {
			sugar.Error(err)
		}
	}

	err = routes.stripeRepo.UpsertInvoice(result.Subs.LatestInvoice)
	if err != nil {
		sugar.Error(err)
	}

	// Payment intent is zero value if subscription is created via introductory offer
	if !result.Subs.PaymentIntent.IsZero() {
		err = routes.stripeRepo.UpsertPaymentIntent(result.Subs.PaymentIntent)
		if err != nil {
			sugar.Error(err)
		}
	}

	if !result.Versioned.IsZero() {
		err := routes.readerRepo.VersionMembership(result.Versioned)
		if err != nil {
			sugar.Error(err)
		}
	}
}

func (routes StripeRoutes) findCartItem(params stripe.SubsParams) (reader.CartItemStripe, error) {
	// Get paywall from cache
	paywall, err := routes.cacheRepo.LoadPaywall(routes.live)
	// If paywall data is found in cache.
	if err == nil {
		item, err := params.BuildCartItem(paywall.Stripe)
		if err == nil {
			return item, nil
		}
	}

	// Cannot build cart item from cached data.
	// Load from db or stripe.
	item, err := routes.stripeRepo.LoadCheckoutItem(params)
	if err != nil {
		return reader.CartItemStripe{}, err
	}

	// In case item in cart is fetched from Stripe API,
	// save it to db.
	if item.AnyFromStripe() {
		go func() {
			defer routes.logger.Sync()
			sugar := routes.logger.Sugar()

			if item.Recurring.IsFromStripe {
				err := routes.stripeRepo.UpsertPrice(item.Recurring)
				if err != nil {
					sugar.Error(err)
				}
			}

			if item.Introductory.IsFromStripe {
				err := routes.stripeRepo.UpsertPrice(item.Introductory)
				if err != nil {
					sugar.Error(err)
				}
			}

			if item.Coupon.IsFromStripe {
				err := routes.stripeRepo.UpsertCoupon(item.Coupon)
				if err != nil {
					sugar.Error(err)
				}
			}
		}()
	}

	return item, nil
}

func (routes StripeRoutes) saveShoppingSession(s stripe.ShoppingSession) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	err := routes.stripeRepo.SaveShoppingSession(s)
	if err != nil {
		sugar.Error(err)
	}

	if s.Subs.ID == "" {
		return
	}

	// Coupon is redeemed only when there's a subscription.
	redeemed := s.CouponRedeemed()
	if redeemed.IsZero() {
		return
	}

	err = routes.stripeRepo.InsertCouponRedeemed(redeemed)
	if err != nil {
		sugar.Error(err)
	}
}
