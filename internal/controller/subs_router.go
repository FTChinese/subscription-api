package controller

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/ftcpay"
	"github.com/FTChinese/subscription-api/internal/repository/products"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"net/http"
)

// SubsRouter is the base type used to handle shared payment operations.
type SubsRouter struct {
	ftcpay.FtcPay // This contains readers.Env to access account data.
	prodRepo      products.Env
	config        config.BuildConfig
}

func NewSubsRouter(dbs db.ReadWriteMyDBs, c *cache.Cache, cfg config.BuildConfig, p postoffice.PostOffice, logger *zap.Logger) SubsRouter {

	return SubsRouter{
		FtcPay:   ftcpay.New(dbs, p, logger),
		prodRepo: products.NewEnv(dbs, c),
		config:   cfg,
	}
}

// Centralized error handling after order creation.
// It handles the errors propagated from Membership.AliWxSubsKind(),
func (router SubsRouter) handleOrderErr(w http.ResponseWriter, err error) {
	var ve *render.ValidationError
	if errors.As(err, &ve) {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	_ = render.New(w).DBError(err)
}

func (router SubsRouter) postOrderCreation(order subs.Order, client footprint.Client) error {
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

func (router SubsRouter) processWebhookResult(result subs.PaymentResult) (subs.ConfirmationResult, *subs.ConfirmError) {
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
		return subs.ConfirmationResult{}, result.ConfirmError(msg, true)
	}

	if !result.IsOrderPaid() {
		return subs.ConfirmationResult{}, result.ConfirmError("order not paid", false)
	}

	order, err := router.SubsRepo.LoadFullOrder(result.OrderID)
	if err != nil {
		sugar.Error(err)
		return subs.ConfirmationResult{}, result.ConfirmError(err.Error(), true)
	}

	return router.ConfirmOrder(result, order)
}
