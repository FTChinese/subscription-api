package api

import (
	"net/http"

	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/products"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
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
// Query parameter:
// ?fresh=<bool>
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

// MigrateToActivePrices performs migration to product_active_price table.
func (router PaywallRouter) MigrateToActivePrices(w http.ResponseWriter, req *http.Request) {
	paywall, err := router.LoadCachedPaywall(false)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	for _, v := range paywall.FTCPrices {
		err = router.productRepo.ActivatePrice(v.FtcPrice)
		if err != nil {
			_ = render.New(w).DBError(err)
			return
		}
	}

	for _, v := range paywall.Stripe {
		err = router.stripeRepo.ActivatePrice(v.Price)
		if err != nil {
			_ = render.New(w).DBError(err)
			return
		}
	}

	_ = render.New(w).NoContent()
}
