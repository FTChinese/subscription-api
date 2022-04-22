package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/products"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"net/http"
)

// PaywallRouter handles pricing plans.
type PaywallRouter struct {
	productRepo products.Env
	PaymentShared
}

func NewPaywallRouter(
	dbs db.ReadWriteMyDBs,
	c *cache.Cache,
	logger *zap.Logger,
	live bool,
) PaywallRouter {
	return PaywallRouter{
		productRepo: products.New(dbs),
		PaymentShared: NewPaymentShared(
			dbs,
			c,
			logger,
			live),
	}
}

// LoadPaywall loads paywall data from db or cache.
func (router PaywallRouter) LoadPaywall(w http.ResponseWriter, req *http.Request) {

	refresh := xhttp.ParseQueryRefresh(req)

	paywall, err := router.LoadCachedPaywall(refresh)

	if err != nil {
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	// Save stripe prices if any of them is fetched from
	// Stripe API.
	go func() {
		defer router.logger.Sync()
		sugar := router.logger.Sugar()

		for _, item := range paywall.Stripe {
			if item.Price.IsFromStripe {
				err := router.stripeRepo.UpsertPrice(item.Price)
				if err != nil {
					sugar.Error(err)
				}
			}

			for _, coupon := range item.Coupons {
				if coupon.IsFromStripe {
					err := router.stripeRepo.UpsertCoupon(coupon)
					if err != nil {
						sugar.Error(err)
					}
				}
			}
		}
	}()

	if refresh {
		go func() {
			defer router.logger.Sync()
			sugar := router.logger.Sugar()

			_, err := router.stripeRepo.ListPricesCompat(router.live, true)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).JSON(http.StatusOK, paywall)
}

func (router PaywallRouter) LoadFtcActivePrices(w http.ResponseWriter, _ *http.Request) {
	p, err := router.paywallRepo.ListActivePrices(router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).JSON(http.StatusOK, p)
}
