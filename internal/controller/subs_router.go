package controller

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/products"
	"github.com/FTChinese/subscription-api/internal/repository/readerrepo"
	"github.com/FTChinese/subscription-api/internal/repository/subrepo"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/client"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"net/http"
)

// SubsRouter is the base type used to handle shared payment operations.
type SubsRouter struct {
	subRepo    subrepo.Env
	readerRepo readerrepo.Env
	prodRepo   products.Env
	postman    postoffice.PostOffice
	config     config.BuildConfig
	logger     *zap.Logger

	aliPayClient subrepo.AliPayClient

	wxPayClients subrepo.WxPayClientStore
}

func NewSubsRouter(db *sqlx.DB, c *cache.Cache, cfg config.BuildConfig, p postoffice.PostOffice, logger *zap.Logger) SubsRouter {

	aliApp := ali.MustInitApp()
	wxApps := wechat.MustGetPayApps()

	return SubsRouter{
		subRepo:      subrepo.NewEnv(db, c, logger),
		readerRepo:   readerrepo.NewEnv(db),
		prodRepo:     products.NewEnv(db, c),
		postman:      p,
		config:       cfg,
		logger:       logger,
		aliPayClient: subrepo.NewAliPayClient(aliApp, logger),
		wxPayClients: subrepo.NewWxClientStore(wxApps, logger),
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

func (router SubsRouter) afterOrderCreated(pi subs.PaymentIntent, client client.Client) error {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	if pi.Kind == enum.OrderKindUpgrade {
		err := router.subRepo.SaveProratedOrders(pi.ProratedOrders())
		if err != nil {
			sugar.Error(err)
			return err
		}
	}

	go func() {
		err := router.subRepo.LogOrderMeta(subs.OrderMeta{
			OrderID: pi.Order.ID,
			Client:  client,
		})
		if err != nil {
			sugar.Error(err)
		}
	}()

	return nil
}

//https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
func (router SubsRouter) verifyWxPayment(order subs.Order) (subs.PaymentResult, error) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	if order.WxAppID.IsZero() {
		order.WxAppID = null.StringFrom(wxAppNativeApp)
	}

	payClient, err := router.wxPayClients.ClientByAppID(order.WxAppID.String)
	if err != nil {
		sugar.Error(err)
		return subs.PaymentResult{}, err
	}

	wxOrder, err := payClient.QueryOrder(order)
	// If there are any errors when querying order.
	if err != nil {
		sugar.Error(err)
		return subs.PaymentResult{}, err
	}

	// Save raw response.
	go func() {
		if err := router.subRepo.SaveWxQueryResp(wxOrder); err != nil {
			sugar.Error(err)
		}
	}()

	// Validate if response is correct. This does not verify the payment is successful.
	// We have to send the payment status as is to client.
	// field: return_code, code: invalid
	// field: result_code, code: invalid
	// field: app_id, code: invalid
	// field: mch_id, code: invalid
	err = wxOrder.Validate(payClient.GetApp())
	if err != nil {
		sugar.Error(err)
		return subs.PaymentResult{}, err
	}

	return subs.NewWxPayResult(wxOrder), nil
}

// https://opendocs.alipay.com/apis/api_1/alipay.trade.query/
func (router SubsRouter) verifyAliPayment(order subs.Order) (subs.PaymentResult, error) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	aliOrder, err := router.aliPayClient.QueryOrder(order.ID)

	if err != nil {
		sugar.Error(err)
		return subs.PaymentResult{}, err
	}

	return subs.NewAliPayResult(aliOrder), nil
}

func (router SubsRouter) processPaymentResult(result subs.PaymentResult) (subs.ConfirmationResult, *subs.ConfirmError) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	confirmed, cfmErr := router.subRepo.ConfirmOrder(result)
	if cfmErr != nil {
		go func() {
			err := router.subRepo.SaveConfirmationErr(cfmErr)
			if err != nil {
				sugar.Error(err)
			}
		}()
		return confirmed, cfmErr
	}

	go func() {
		if !confirmed.Snapshot.IsZero() {
			err := router.readerRepo.BackUpMember(confirmed.Snapshot)
			if err != nil {
				sugar.Error(err)
			}
		}

		// Flag upgrade balance as consumed.
		if confirmed.Order.Kind == enum.OrderKindUpgrade {
			err := router.subRepo.ProratedOrdersUsed(confirmed.Order.ID)
			if err != nil {
				sugar.Error(err)
			}
		}

		if err := router.sendConfirmationEmail(confirmed.Order); err != nil {
			sugar.Error(err)
		}
	}()

	return confirmed, nil
}

// SendConfirmationLetter sends a confirmation email if user logged in with FTC account.
func (router SubsRouter) sendConfirmationEmail(order subs.Order) error {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// If the FtcID field is null, it indicates this user
	// does not have an FTC account linked. You cannot find out
	// its email address.
	if !order.FtcID.Valid {
		return nil
	}
	// Find this user's personal data
	account, err := router.readerRepo.AccountByFtcID(order.FtcID.String)

	if err != nil {
		return err
	}

	var parcel postoffice.Parcel
	switch order.Kind {
	case enum.OrderKindCreate:
		parcel, err = letter.NewSubParcel(account, order)

	case enum.OrderKindRenew:
		parcel, err = letter.NewRenewalParcel(account, order)

	case enum.OrderKindUpgrade:
		pos, err := router.subRepo.ListProratedOrders(order.ID)
		if err != nil {
			return err
		}
		parcel, err = letter.NewUpgradeParcel(account, order, pos)
	}

	if err != nil {
		sugar.Error(err)
		return err
	}

	sugar.Info("Send subscription confirmation letter")

	err = router.postman.Deliver(parcel)
	if err != nil {
		sugar.Error(err)
		return err
	}
	return nil
}
