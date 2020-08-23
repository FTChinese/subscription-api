package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/go-rest/view"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/client"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/sirupsen/logrus"
	"github.com/smartwalle/alipay"
	"go.uber.org/zap"
	"net/http"
)

const (
	success = "success"
	fail    = "fail"
)

// AliPayRouter handles alipay request
type AliPayRouter struct {
	appID  string
	client *alipay.AliPay
	PayRouter
}

// NewAliRouter create a new instance of AliPayRouter
func NewAliRouter(baseRouter PayRouter) AliPayRouter {

	app := ali.MustInitApp()

	r := AliPayRouter{
		appID:     app.ID,
		client:    alipay.New(app.ID, app.PublicKey, app.PrivateKey, true),
		PayRouter: baseRouter,
	}

	return r
}

// PlaceOrder creates an http handler function depending
// on the device platform.
//
// 	POST /<desktop|mobile|app>/{tier}/{cycle}?<return_url=xxx>
// `return_url` parameter is only required for apps running on ftacademy.cn
func (router AliPayRouter) PlaceOrder(kind ali.EntryKind) http.HandlerFunc {

	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	sugar.Infow("Create wxpay order",
		"trace", "WxPayRouter.PlaceOrder",
		"platform", kind.String(),
	)

	// TODO: put all those fields in request body.
	// returnUrl: string;
	// tier: string;
	// cycle: string;
	// planId: string;
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			sugar.Error(err)
			_ = view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		clientApp := client.NewClientApp(req)
		readerIDs := getReaderIDs(req.Header)

		input, err := gatherAliPayInput(req)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).BadRequest(err.Error())
			return
		}
		if ve := input.Validate(); ve != nil {
			sugar.Error(err)
			_ = render.New(w).Unprocessable(ve)
			return
		}

		plan, err := router.prodRepo.PlanByEdition(input.Edition)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).BadRequest(err.Error())
			return
		}

		builder := subs.NewOrderBuilder(readerIDs).
			SetPlan(plan).
			SetPayMethod(enum.PayMethodAli)

		order, err := router.subEnv.CreateOrder(builder)

		if err != nil {
			sugar.Error(err)
			router.handleOrderErr(w, err)
			return
		}

		sugar.Infof("Created order: %+v", order)

		go func() {
			_ = router.subEnv.SaveOrderClient(client.OrderClient{
				OrderID: order.ID,
				Client:  clientApp,
			})
		}()

		switch kind {
		case ali.EntryApp:
			// Generate signature and creates a string of query
			// parameter.
			queryStr, err := router.client.TradeAppPay(
				builder.AliAppPayParams())

			sugar.Infof("App pay param: %+v\n", queryStr)

			if err != nil {
				sugar.Error(err)
				_ = render.New(w).BadRequest(err.Error())
				return
			}

			_ = render.New(w).
				JSON(http.StatusOK, subs.AlipayNativeAppOrder{
					Order: order,
					Param: queryStr,
				})
			return

		case ali.EntryDesktopWeb:
			redirectURL, err := router.client.TradePagePay(
				builder.AliDesktopPayParams(input.ReturnURL),
			)
			if err != nil {
				sugar.Error(err)
				_ = render.New(w).BadRequest(err.Error())
				return
			}

			sugar.Infof("Ali desktop browser redirect url: %+v\n", redirectURL)

			_ = render.New(w).JSON(http.StatusOK, subs.AlipayBrowserOrder{
				Order:       order,
				RedirectURL: redirectURL.String(),
			})

		case ali.EntryMobileWeb:
			redirectURL, err := router.client.TradeWapPay(
				builder.AliWapPayParams(input.ReturnURL),
			)
			if err != nil {
				sugar.Error(err)
				_ = render.New(w).BadRequest(err.Error())
				return
			}
			sugar.Infof("Ali mobile browser redirect url: %+v\n", redirectURL)

			_ = render.New(w).JSON(http.StatusOK, subs.AlipayBrowserOrder{
				Order:       order,
				RedirectURL: input.ReturnURL,
			})
		}
	}
}

// WebHook handles alipay server-side notification.
func (router AliPayRouter) WebHook(w http.ResponseWriter, req *http.Request) {
	logger := logrus.WithFields(logrus.Fields{
		"trace": "AliPayRouter.WebHook",
	})

	err := req.ParseForm()
	if err != nil {
		logger.Error(err)

		if _, err := w.Write([]byte(fail)); err != nil {
			logger.Error(err)
		}
		return
	}

	// If err is nil, then the signature is verified.
	noti, err := router.client.GetTradeNotification(req)
	logger.WithFields(logrus.Fields{
		"event": "NotificationBody",
	}).Infof("+%v", noti)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"event": "GetTradeNotification",
		}).Error(err)

		if _, err := w.Write([]byte(fail)); err != nil {
			logger.Error(err)
		}
		return
	}

	// 4、验证app_id是否为该商户本身
	if noti.AppId != router.appID {
		logger.WithFields(logrus.Fields{
			"event":   "AppIDNotMatch",
			"orderId": noti.OutTradeNo,
		}).Infof("Expected %s, actual %s", router.appID, noti.AppId)

		if _, err := w.Write([]byte(fail)); err != nil {
			logger.Error(err)
		}
		return
	}

	go func() {
		if err := router.subEnv.SaveAliNotification(*noti); err != nil {
			logger.Error(err)
		}
	}()

	// 在支付宝的业务通知中，只有交易通知状态为TRADE_SUCCESS或TRADE_FINISHED时，支付宝才会认定为买家付款成功。
	if !ali.IsPaySuccess(noti) {
		logger.WithFields(logrus.Fields{
			"event":   "PaymentFailed",
			"orderId": noti.OutTradeNo,
		}).Infof("Status %s", noti.TradeStatus)

		if ali.ShouldRetry(noti) {
			if _, err := w.Write([]byte(fail)); err != nil {
				logger.Error(err)
			}
			return
		}

		if _, err := w.Write([]byte(success)); err != nil {
			logger.Error(err)
		}
		return
	}

	payResult, err := subs.NewPaymentResultAli(noti)

	// 1、商户需要验证该通知数据中的out_trade_no是否为商户系统中创建的订单号
	// 2、判断total_amount是否确实为该订单的实际金额（即商户订单创建时的金额）
	confirmed, confirmErr := router.subEnv.ConfirmOrder(payResult)

	if confirmErr != nil {

		go func() {
			_ = router.subEnv.SaveConfirmationResult(
				confirmErr.Schema(payResult.OrderID),
			)
		}()

		if confirmErr.Retry {
			if b, err := w.Write([]byte(fail)); err != nil {
				logger.WithField("byte_output", b).Error(err)
			}
			return
		} else {
			if b, err := w.Write([]byte(success)); err != nil {
				logger.WithField("byte_output", b).Error(err)
			}
			return
		}
	}

	if !confirmed.Snapshot.IsZero() {
		_ = router.readerEnv.BackUpMember(confirmed.Snapshot)
	}

	go func() {
		if err := router.sendConfirmationEmail(confirmed.Order); err != nil {
			logger.Error(err)
		}
	}()

	if b, err := w.Write([]byte(success)); err != nil {
		logger.WithField("byte_output", b).Error(err)
	}
}
