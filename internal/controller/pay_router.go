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
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
	"github.com/objcoding/wxpay"
	"github.com/patrickmn/go-cache"
	"github.com/smartwalle/alipay"
	"net/http"
)

// PayRouter is the base type used to handle shared payment operations.
type PayRouter struct {
	subRepo    subrepo.Env
	readerRepo readerrepo.Env
	prodRepo   products.Env
	postman    postoffice.PostOffice
	config     config.BuildConfig

	aliAppID string
	aliPay   *alipay.AliPay

	wxPayClients wechat.PayClients
}

func NewPayRouter(db *sqlx.DB, c *cache.Cache, b config.BuildConfig, p postoffice.PostOffice) PayRouter {

	aliApp := ali.MustInitApp()

	return PayRouter{
		subRepo:    subrepo.NewEnv(db, c),
		readerRepo: readerrepo.NewEnv(db),
		prodRepo:   products.NewEnv(db, c),
		postman:    p,
		config:     b,

		aliAppID: aliApp.ID,
		aliPay:   alipay.New(aliApp.ID, aliApp.PublicKey, aliApp.PrivateKey, true),

		wxPayClients: wechat.InitPayClients(),
	}
}

// Only when the user has ftc account, and
// query parameter has `test=true` will
// we search db to see whether it is actually
// a test account.
func (router PayRouter) isTestAccount(ids reader.MemberID, req *http.Request) bool {
	isTest := ids.FtcID.Valid && req.FormValue("test") == "true"

	if !isTest {
		return false
	}

	found, err := router.readerRepo.SandboxUserExists(ids.FtcID.String)
	if err != nil {
		return false
	}

	return found
}

// Centralized error handling after order creation.
// It handles the errors propagated from Membership.AliWxSubsKind(),
func (router PayRouter) handleOrderErr(w http.ResponseWriter, err error) {
	var ve *render.ValidationError
	if errors.As(err, &ve) {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	_ = render.New(w).DBError(err)
}

// SendConfirmationLetter sends a confirmation email if user logged in with FTC account.
func (router PayRouter) sendConfirmationEmail(order subs.Order) error {
	defer logger.Sync()
	sugar := logger.Sugar()

	// If the FtcID field is null, it indicates this user
	// does not have an FTC account linked. You cannot find out
	// its email address.
	if !order.FtcID.Valid {
		return nil
	}
	// Find this user's personal data
	account, err := router.readerRepo.FtcAccountByFtcID(order.FtcID.String)

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

func (router PayRouter) queryWxOrder(order subs.Order) (subs.PaymentResult, *render.ResponseError) {
	defer logger.Sync()
	sugar := logger.Sugar()

	if order.WxAppID.IsZero() {
		order.WxAppID = null.StringFrom(wxAppNativeApp)
	}

	payClient, err := router.wxPayClients.GetClientByAppID(order.WxAppID.String)
	if err != nil {
		sugar.Error(err)
		return subs.PaymentResult{}, render.NewInternalError(err.Error())
	}

	reqParams := make(wxpay.Params)
	reqParams.SetString("out_trade_no", order.ID)

	// Send query to Wechat server
	// Returns the parsed response as a map.
	// It checks if the response contains `return_code` key.
	// If return_code == FAIL, it does not returns error.
	// If return_code == SUCCESS, it verifies the signature.
	respParams, err := payClient.OrderQuery(reqParams)

	// If there are any errors when querying order.
	if err != nil {
		sugar.Error(err)
		return subs.PaymentResult{}, render.NewInternalError(err.Error())
	}

	sugar.Infof("Wechat order found")

	// Response:
	// {message: "", {field: status, code: fail} }
	// {message: "", {field: result, code: "ORDERNOTEXIST" | "SYSTEMERROR"} }
	resp := wechat.NewOrderQueryResp(respParams)
	go func() {
		if err := router.subRepo.SaveWxQueryResp(resp); err != nil {
			sugar.Error(err)
		}
	}()

	if r := resp.Validate(payClient.GetApp()); r != nil {
		sugar.Info("Response invalid")

		if r.Field == "result" && r.Code == "ORDERNOTEXIST" {
			return subs.PaymentResult{}, render.NewNotFound("Order does not exist")
		}

		return subs.PaymentResult{}, render.NewUnprocessable(r)
	}

	return subs.NewWxQueryResult(resp), nil
}

func (router PayRouter) queryAliOrder(order subs.Order) (subs.PaymentResult, *render.ResponseError) {
	result, err := router.aliPay.TradeQuery(alipay.AliPayTradeQuery{
		AppAuthToken: "",
		OutTradeNo:   order.ID,
	})

	if err != nil {
		return subs.PaymentResult{}, render.NewBadRequest(err.Error())
	}

	if !result.IsSuccess() {
		return subs.PaymentResult{}, render.NewUnprocessable(&render.ValidationError{
			Message: result.AliPayTradeQuery.Msg,
			Field:   "status",
			Code:    render.InvalidCode(result.AliPayTradeQuery.TradeStatus),
		})
	}

	return subs.NewAliQueryResult(result), nil
}
