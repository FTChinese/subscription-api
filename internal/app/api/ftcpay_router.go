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

// FtcPayRoutes is the base type used to handle shared payment operations.
type FtcPayRoutes struct {
	paybase.FtcPayBase // This contains readers.Env to access account data.
	paywallRepo        repository.PaywallRepo
	cacheRepo          repository.CacheRepo
	live               bool // Determine webhook url. If true, use production server; otherwise goes to sandbox server.
}

func NewFtcPayRoutes(
	dbs db.ReadWriteMyDBs,
	c *cache.Cache,
	logger *zap.Logger,
	live bool,
) FtcPayRoutes {
	return FtcPayRoutes{
		FtcPayBase:  paybase.NewFtcPay(dbs, logger),
		paywallRepo: repository.NewPaywallRepo(dbs),
		cacheRepo:   repository.NewCacheRepo(c),
		live:        live,
	}
}

// Centralized error handling after order creation.
// It handles the errors propagated from Membership.AliWxSubsKind(),
func (routes FtcPayRoutes) handleOrderErr(w http.ResponseWriter, err error) {
	var ve *render.ValidationError
	if errors.As(err, &ve) {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	_ = render.New(w).DBError(err)
}

func (routes FtcPayRoutes) postOrderCreation(order ftcpay.Order, client footprint.Client) error {
	defer routes.Logger.Sync()
	sugar := routes.Logger.Sugar()

	go func() {
		err := routes.SubsRepo.SaveOrderMeta(footprint.OrderClient{
			OrderID: order.ID,
			Client:  client,
		})
		if err != nil {
			sugar.Error(err)
		}
	}()

	return nil
}

func (routes FtcPayRoutes) processWebhookResult(result ftcpay.PaymentResult) (ftcpay.ConfirmationResult, *ftcpay.ConfirmError) {
	defer routes.Logger.Sync()
	sugar := routes.Logger.Sugar()

	sugar.Infof("Webhook Payment result %v", result)

	go func() {
		err := routes.SubsRepo.SavePayResult(result)
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

	order, err := routes.SubsRepo.LoadFullOrder(result.OrderID)
	if err != nil {
		sugar.Error(err)
		return ftcpay.ConfirmationResult{}, result.ConfirmError(err.Error(), true)
	}

	return routes.ConfirmOrder(result, order)
}

func (routes FtcPayRoutes) loadCheckoutItem(params ftcpay.FtcCartParams, live bool) (reader.CartItemFtc, *render.ResponseError) {
	defer routes.Logger.Sync()
	sugar := routes.Logger.Sugar()

	sugar.Infof("Load checkout item. Live %t", live)

	paywall, err := routes.cacheRepo.LoadPaywall(live)
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
	ci, err := routes.paywallRepo.LoadCheckoutItem(
		params,
		routes.live)

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
