package controller

import (
	"errors"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/ftcpay"
	"github.com/FTChinese/subscription-api/internal/repository/products"
	"github.com/FTChinese/subscription-api/pkg/client"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"net/http"
)

// SubsRouter is the base type used to handle shared payment operations.
type SubsRouter struct {
	ftcpay.FtcPay
	prodRepo products.Env
	config   config.BuildConfig
}

func NewSubsRouter(db *sqlx.DB, c *cache.Cache, cfg config.BuildConfig, p postoffice.PostOffice, logger *zap.Logger) SubsRouter {

	return SubsRouter{
		FtcPay:   ftcpay.New(db, p, logger),
		prodRepo: products.NewEnv(db, c),
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

func (router SubsRouter) postOrderCreation(order subs.Order, client client.Client) error {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	go func() {
		err := router.SubsRepo.LogOrderMeta(subs.OrderMeta{
			OrderID: order.ID,
			Client:  client,
		})
		if err != nil {
			sugar.Error(err)
		}
	}()

	return nil
}
