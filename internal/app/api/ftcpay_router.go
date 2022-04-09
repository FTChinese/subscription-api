package api

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/app/paybase"
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/internal/repository"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"net/http"
)

// FtcPayRouter is the base type used to handle shared payment operations.
type FtcPayRouter struct {
	paybase.FtcPayBase // This contains readers.Env to access account data.
	paywallRepo        repository.PaywallRepo
	cacheRepo          repository.CacheRepo
	live               bool // Determine webhook url. If true, use production server; otherwise goes to sandbox server.
}

func NewFtcPayRouter(
	dbs db.ReadWriteMyDBs,
	c *cache.Cache,
	logger *zap.Logger,
	live bool,
) FtcPayRouter {
	return FtcPayRouter{
		FtcPayBase:  paybase.NewFtcPay(dbs, logger),
		paywallRepo: repository.NewPaywallRepo(dbs),
		cacheRepo:   repository.NewCacheRepo(c),
		live:        live,
	}
}

// Centralized error handling after order creation.
// It handles the errors propagated from Membership.AliWxSubsKind(),
func (router FtcPayRouter) handleOrderErr(w http.ResponseWriter, err error) {
	var ve *render.ValidationError
	if errors.As(err, &ve) {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	_ = render.New(w).DBError(err)
}

func (router FtcPayRouter) postOrderCreation(order ftcpay.Order, client footprint.Client) error {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	go func() {
		err := router.SubsRepo.SaveOrderMeta(footprint.OrderClient{
			OrderID: order.ID,
			Client:  client,
		})
		if err != nil {
			sugar.Error(err)
		}
	}()

	return nil
}

func (router FtcPayRouter) processWebhookResult(result ftcpay.PaymentResult) (ftcpay.ConfirmationResult, *ftcpay.ConfirmError) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	sugar.Infof("Webhook Payment result %v", result)

	go func() {
		err := router.SubsRepo.SavePayResult(result)
		if err != nil {
			sugar.Error(err)
		}
	}()

	if result.ShouldRetry() {
		msg := fmt.Sprintf("payment status %s", result.PaymentState)
		return ftcpay.ConfirmationResult{}, result.ConfirmError(msg, true)
	}

	if !result.IsOrderPaid() {
		return ftcpay.ConfirmationResult{}, result.ConfirmError("order not paid", false)
	}

	order, err := router.SubsRepo.LoadFullOrder(result.OrderID)
	if err != nil {
		sugar.Error(err)
		return ftcpay.ConfirmationResult{}, result.ConfirmError(err.Error(), true)
	}

	return router.ConfirmOrder(result, order)
}

func (router FtcPayRouter) loadCheckoutItem(params ftcpay.FtcCartParams, live bool) (reader.CartItemFtc, *render.ResponseError) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	sugar.Infof("Load checkout item. Live %t", live)

	paywall, err := router.cacheRepo.LoadPaywall(live)
	// If price and discount could be found in paywall.
	if err == nil {
		sugar.Infof("Paywall Cache found. Search checkout item.")
		item, err := params.BuildCartItem(paywall.Products)
		if err == nil {
			sugar.Infof("Checkout item found in cache")
			return item, nil
		} else {
			sugar.Infof("Checkout item not found in cache. Search db directly.")
		}
	}

	// Otherwise, retrieve from db.
	ci, err := router.paywallRepo.LoadCheckoutItem(
		params,
		router.live)

	if err != nil {
		sugar.Error(err)
		return reader.CartItemFtc{}, render.NewDBError(err)
	}

	if err := ci.Verify(live); err != nil {
		sugar.Error(err)
		return reader.CartItemFtc{}, render.NewBadRequest(err.Error())
	}

	return ci, nil
}
