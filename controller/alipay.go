package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/sirupsen/logrus"
	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/models"
	"gitlab.com/ftchinese/subscription-api/models/ali"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository/subrepo"
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
func NewAliRouter(subEnv subrepo.SubEnv, p postoffice.Postman) AliPayRouter {

	app := getAliPayApp()

	client := alipay.New(app.ID, app.PublicKey, app.PrivateKey, true)

	r := AliPayRouter{
		appID:  app.ID,
		client: client,
	}
	r.subEnv = subEnv
	r.postman = p

	return r
}

// PlaceOrder creates an http handler function depending
// on the device platform.
//
// 	POST /<desktop|mobile|app>/{tier}/{cycle}?<return_url=xxx>
// `return_url` parameter is only required for apps running on ftacademy.cn
func (router AliPayRouter) PlaceOrder(kind ali.EntryKind) http.HandlerFunc {
	logger := logrus.WithFields(logrus.Fields{
		"trace": "AliPayRouter.PlaceOrder",
		"type":  kind.String(),
	})

	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			logger.Error(err)
			_ = view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		clientApp := util.NewClientApp(req)

		logger.Infof("Client app: %+v", clientApp)

		if err := models.AllowAndroidPurchase(clientApp); err != nil {
			logger.Error(err)
			_ = view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		userID, _ := GetUserID(req.Header)

		plan, err := router.findPlan(req)

		if err != nil {
			logger.Error(err)

			_ = view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		builder := subscription.NewOrderBuilder(userID).
			SetPlan(plan).
			SetPayMethod(enum.PayMethodAli)

		// TODO: save client info
		order, err := router.subEnv.CreateOrder(builder)

		//order, err := router.createOrder(
		//	userID,
		//	plan,
		//	enum.PayMethodAli,
		//	clientApp,
		//	null.String{},
		//)
		//
		if err != nil {
			logger.Error(err)
			router.handleOrderErr(w, err)
			return
		}

		logger.Infof("Created order: %+v", order)

		go func() {
			_ = router.subEnv.SaveOrderClient(builder.ClientApp())
		}()

		// Alipay specific handling.
		returnURL := req.FormValue("return_url")

		switch kind {
		case ali.EntryApp:
			// Generate signature and creates a string of query
			// parameter.
			queryStr, err := router.client.TradeAppPay(
				builder.AliAppPayParams())

			logger.Infof("App pay param: %+v\n", queryStr)

			if err != nil {
				logger.Error(err)
				_ = view.Render(w, view.NewBadRequest(err.Error()))
				return
			}

			_ = view.Render(
				w,
				view.NewResponse().SetBody(subscription.AlipayNativeAppOrder{
					Order: order,
					Param: queryStr,
				}),
			)
			return

		case ali.EntryDesktopWeb:
			redirectURL, err := router.client.TradePagePay(
				builder.AliDesktopPayParams(returnURL),
			)
			if err != nil {
				logger.Error(err)
				_ = view.Render(w, view.NewBadRequest(err.Error()))
				return
			}

			logger.Infof("Ali desktop browser redirect url: %+v\n", redirectURL)

			_ = view.Render(w, view.NewResponse().SetBody(subscription.AlipayBrowserOrder{
				Order: order,
				RedirectURL: redirectURL.String(),
			}))

		case ali.EntryMobileWeb:
			redirectURL, err := router.client.TradeWapPay(
				builder.AliWapPayParams(returnURL),
			)
			if err != nil {
				logger.Error(err)
				_ = view.Render(w, view.NewBadRequest(err.Error()))
				return
			}
			logger.Infof("Ali mobile browser redirect url: %+v\n", redirectURL)

			_ = view.Render(w, view.NewResponse().SetBody(subscription.AlipayBrowserOrder{
				Order: order,
				RedirectURL:returnURL,
			}))
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

	payResult, err := ali.GetPaymentResult(noti)

	// 1、商户需要验证该通知数据中的out_trade_no是否为商户系统中创建的订单号
	// 2、判断total_amount是否确实为该订单的实际金额（即商户订单创建时的金额）
	confirmedOrder, result := router.confirmPayment(payResult)

	if result != nil {
		logger.Error(err)

		go func() {
			err := router.subEnv.SaveConfirmationResult(result)
			if err != nil {
				logger.Error(err)
			}
		}()

		if result.Retry {
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

	go func() {
		if err := router.sendConfirmationEmail(confirmedOrder); err != nil {
			logger.Error(err)
		}
	}()

	go func() {
		if err := router.subEnv.SaveConfirmationResult(subscription.NewConfirmationSucceeded(payResult.OrderID)); err != nil {
			logger.Error(err)
		}
	}()

	if b, err := w.Write([]byte(success)); err != nil {
		logger.WithField("byte_output", b).Error(err)
	}
}
