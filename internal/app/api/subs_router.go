package api

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/app/paybase"
	"github.com/FTChinese/subscription-api/internal/pkg/subs"
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"net/http"
)

// SubsRouter is the base type used to handle shared payment operations.
type SubsRouter struct {
	paybase.FtcPayBase // This contains readers.Env to access account data.
	PaywallRepo        shared.PaywallCommon
	Live               bool // Determine webhook url. If true, use production server; otherwise goes to sandbox server.
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

func (router SubsRouter) loadCheckoutItem(params pw.CartParams, live bool) (price.CheckoutItem, *render.ResponseError) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	sugar.Infof("Load checkout item. Live %t", live)

	paywall, err := router.PaywallRepo.LoadPaywall(live)
	// If price and discount could be found in paywall.
	if err == nil {
		sugar.Infof("Paywall Cache found. Search checkout item.")
		item, err := paywall.FindCheckoutItem(params)
		if err == nil {
			sugar.Infof("Checkout item found in cache")
			return item, nil
		} else {
			sugar.Infof("Checkout item not found in cache. Search db directly.")
		}
	}

	// Otherwise, retrieve from db.
	ci, err := router.PaywallRepo.LoadCheckoutItem(
		params,
		router.Live)

	if err != nil {
		sugar.Error(err)
		return price.CheckoutItem{}, render.NewDBError(err)
	}

	if err := ci.Verify(live); err != nil {
		sugar.Error(err)
		return price.CheckoutItem{}, render.NewBadRequest(err.Error())
	}

	return ci, nil
}