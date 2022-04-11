package api

import (
	"github.com/FTChinese/subscription-api/internal/repository"
	"github.com/FTChinese/subscription-api/internal/repository/stripeenv"
	"github.com/FTChinese/subscription-api/internal/stripeclient"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

type PaymentShared struct {
	stripeRepo  stripeenv.Env
	paywallRepo repository.PaywallRepo
	cacheRepo   repository.CacheRepo
	logger      *zap.Logger
	live        bool
}

func NewPaymentShared(
	dbs db.ReadWriteMyDBs,
	c *cache.Cache,
	logger *zap.Logger,
	live bool,
) PaymentShared {
	return PaymentShared{
		stripeRepo: stripeenv.New(
			stripeclient.New(live, logger),
			repository.NewStripeRepo(dbs, logger),
		),
		paywallRepo: repository.NewPaywallRepo(dbs),
		cacheRepo:   repository.NewCacheRepo(c),
		logger:      logger,
		live:        live,
	}
}

// LoadPaywall directly from database or Stripe API.
func (ps PaymentShared) LoadPaywall() (reader.Paywall, error) {
	defer ps.logger.Sync()
	sugar := ps.logger.Sugar()

	paywall, err := ps.paywallRepo.RetrievePaywall(ps.live)
	if err != nil {
		sugar.Error(err)
		return reader.Paywall{}, err
	}

	stripeIDs := paywall.StripePriceIDs()
	stripePrices, err := ps.stripeRepo.ListPrices(stripeIDs)
	if err != nil {
		return reader.Paywall{}, err
	}

	paywall.StripePrices = stripePrices

	return paywall, nil
}

func (ps PaymentShared) LoadCachedPaywall(refresh bool) (reader.Paywall, error) {
	if !refresh {
		paywall, err := ps.cacheRepo.LoadPaywall(ps.live)
		if err == nil {
			return paywall, nil
		}
	}

	paywall, err := ps.LoadPaywall()
	if err != nil {
		return reader.Paywall{}, err
	}

	ps.cacheRepo.CachePaywall(paywall, ps.live)

	return paywall, nil
}
