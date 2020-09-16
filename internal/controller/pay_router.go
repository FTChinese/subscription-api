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
	"go.uber.org/zap"
	"net/http"
)

// PayRouter is the base type used to handle shared payment operations.
type PayRouter struct {
	subRepo    subrepo.Env
	readerRepo readerrepo.Env
	prodRepo   products.Env
	postman    postoffice.PostOffice
	config     config.BuildConfig
	logger     *zap.Logger

	aliAppID string
	aliPay   *alipay.AliPay

	wxPayClients wechat.PayClients
}

func NewPayRouter(db *sqlx.DB, c *cache.Cache, b config.BuildConfig, p postoffice.PostOffice, logger *zap.Logger) PayRouter {

	aliApp := ali.MustInitApp()

	return PayRouter{
		subRepo:    subrepo.NewEnv(db, c, logger),
		readerRepo: readerrepo.NewEnv(db),
		prodRepo:   products.NewEnv(db, c),
		postman:    p,
		config:     b,
		logger:     logger,

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
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

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

//https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_2&index=4
func (router PayRouter) queryWxOrder(order subs.Order) (subs.PaymentResult, *render.ResponseError) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

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
	// Example:
	// appid:***REMOVED***
	// device_info: mch_id:1504993271
	// nonce_str:9dmEFWFU5ooB9dMN
	// out_trade_no:FT9F67C5CC9F47CF65
	// result_code:SUCCESS
	// return_code:SUCCESS
	// return_msg:OK
	// sign:538529EAEE06FE61ECE379C699437B37
	// total_fee:25800
	// trade_state:NOTPAY
	// trade_state_desc:订单未支付
	respParams, err := payClient.OrderQuery(reqParams)

	// If there are any errors when querying order.
	if err != nil {
		sugar.Error(err)
		return subs.PaymentResult{}, render.NewInternalError(err.Error())
	}

	sugar.Infof("Wechat order query result: %+v", respParams)

	resp := wechat.NewOrderQueryResp(respParams)
	go func() {
		if err := router.subRepo.SaveWxQueryResp(resp); err != nil {
			sugar.Error(err)
		}
	}()

	// Validate if response is correct. This does not verify the payment is successful.
	// field: return_code, code: invalid
	// field: result_code, code: invalid
	// field: app_id, code: invalid
	// field: mch_id, code: invalid
	if ve := resp.Validate(payClient.GetApp()); ve != nil {
		if resp.IsOrderNotFound() {
			return subs.PaymentResult{}, render.NewNotFound("Order does not exist")
		}

		return subs.PaymentResult{}, render.NewUnprocessable(ve)
	}

	if ve := resp.ValidateTradeState(); ve != nil {
		return subs.PaymentResult{}, render.NewUnprocessable(ve)
	}

	return subs.NewWxQueryResult(resp), nil
}

// https://opendocs.alipay.com/apis/api_1/alipay.trade.query/
func (router PayRouter) queryAliOrder(order subs.Order) (subs.PaymentResult, *render.ResponseError) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	result, err := router.aliPay.TradeQuery(alipay.AliPayTradeQuery{
		AppAuthToken: "",
		OutTradeNo:   order.ID,
	})

	if err != nil {
		sugar.Error(err)
		return subs.PaymentResult{}, render.NewBadRequest(err.Error())
	}

	sugar.Infof("Alipay trade query result: %+v", result)

	// Example failure response:
	// Code:40004  // 网关返回码
	// Msg:Business Failed // 网关返回码描述
	// SubCode:ACQ.TRADE_NOT_EXIST // 业务返回码
	// SubMsg:交易不存在 // 业务返回码描述
	// AuthTradePayMode:
	// BuyerLogonId: BuyerPayAmount:0.00
	// BuyerUserId:
	// BuyerUserType:
	// InvoiceAmount:0.00
	// OutTradeNo:FT33381C4D23AE4F19
	// PointAmount:0.00 ReceiptAmount:0.00
	// SendPayDate:
	// TotalAmount:
	// TradeNo:
	// TradeStatus: // 交易状态：WAIT_BUYER_PAY（交易创建，等待买家付款）、TRADE_CLOSED（未付款交易超时关闭，或支付完成后全额退款）、TRADE_SUCCESS（交易支付成功）、TRADE_FINISHED（交易结束，不可退款）
	// DiscountAmount:
	// FundBillList:[]
	// MdiscountAmount:
	// PayAmount:
	// PayCurrency:
	// SettleAmount:
	// SettleCurrency:
	// SettleTransRate:
	// StoreId: StoreName:
	// TerminalId: TransCurrency:
	// TransPayRate:
	// DiscountGoodsDetail:
	// IndustrySepcDetail:
	// VoucherDetailList:[]
	if !result.IsSuccess() {
		return subs.PaymentResult{}, render.NewUnprocessable(&render.ValidationError{
			Message: result.AliPayTradeQuery.Msg + ":" + result.AliPayTradeQuery.SubMsg,
			Field:   "code",
			Code:    render.CodeInvalid,
		})
	}

	// 交易状态：
	// WAIT_BUYER_PAY（交易创建，等待买家付款）、
	// TRADE_CLOSED（未付款交易超时关闭，或支付完成后全额退款）、
	// TRADE_SUCCESS（交易支付成功）、
	// TRADE_FINISHED（交易结束，不可退款）
	// We must ensure trade_status is successful here.
	// Golang cannot parse empty string to number. If you
	// don't check it here, the next step will fail.
	if !ali.IsStatusSuccess(result.AliPayTradeQuery.TradeStatus) {
		return subs.PaymentResult{}, render.NewUnprocessable(&render.ValidationError{
			Message: result.AliPayTradeQuery.TradeStatus,
			Field:   "trade_status",
			Code:    render.CodeInvalid,
		})
	}

	return subs.NewAliQueryResult(result), nil
}
