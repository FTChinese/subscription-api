package controller

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/guregu/null"
	"github.com/sirupsen/logrus"
	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/ali"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
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
func NewAliRouter(m model.Env, p postoffice.Postman) AliPayRouter {

	app := getAliPayApp()

	client := alipay.New(app.ID, app.PublicKey, app.PrivateKey, true)

	r := AliPayRouter{
		appID:  app.ID,
		client: client,
	}
	r.model = m
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
			view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		user, _ := GetUserID(req.Header)

		plan, err := router.findPlan(req)

		if err != nil {
			logger.Error(err)

			view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		// Save the subscription
		clientApp := util.NewClientApp(req)

		subs, err := router.model.CreateOrder(
			user,
			plan,
			enum.PayMethodAli,
			clientApp,
			null.String{},
		)
		if err != nil {
			router.handleOrderErr(w, err)
			return
		}

		logger.Infof("Created order: %+v", subs)

		// Alipay specific handling.
		returnURL := req.FormValue("return_url")

		tradePay := alipay.TradePay{
			NotifyURL:   router.aliCallbackURL(),
			ReturnURL:   returnURL,
			Subject:     plan.Title,
			OutTradeNo:  subs.ID,
			TotalAmount: subs.AliPrice(),
		}

		switch kind {
		case ali.EntryApp:
			param := ali.BuildAppPay(tradePay)
			// Generate signature and creates a string of query
			// parameter.
			queryStr, err := router.client.TradeAppPay(param)

			logger.Infof("App pay param: %+v\n", queryStr)

			if err != nil {
				logger.Error(err)
				view.Render(w, view.NewBadRequest(err.Error()))
				return
			}
			//resp := ali.NewAppPayResp(subs, queryStr)

			order := ali.NewAppOrder(subs, queryStr)
			view.Render(w, view.NewResponse().SetBody(order))
			// For pay from app you should stop here.
			return

		case ali.EntryDesktopWeb:
			param := ali.BuildDesktopPay(tradePay)
			redirectURL, err := router.client.TradePagePay(param)
			if err != nil {
				logger.Error(err)
				view.Render(w, view.NewBadRequest(err.Error()))
				return
			}

			logger.Infof("Ali desktop browser redirect url: %+v\n", redirectURL)
			order := ali.NewBrowserOrder(subs, redirectURL.String())
			view.Render(w, view.NewResponse().SetBody(order))

		case ali.EntryMobileWeb:
			param := ali.BuildWapPay(tradePay)
			redirectURL, err := router.client.TradeWapPay(param)
			if err != nil {
				logger.Error(err)
				view.Render(w, view.NewBadRequest(err.Error()))
				return
			}
			logger.Infof("Ali mobile browser redirect url: %+v\n", redirectURL)
			order := ali.NewBrowserOrder(subs, redirectURL.String())
			view.Render(w, view.NewResponse().SetBody(order))
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
		if err := router.model.SaveAliNotification(*noti); err != nil {
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

	orderID := noti.OutTradeNo
	// 1、商户需要验证该通知数据中的out_trade_no是否为商户系统中创建的订单号
	charge, err := router.model.FindSubsCharge(orderID)

	// If the order does not exist, tell ali success;
	// If err is not `not found`, tell ali to resend.
	if err != nil {
		logger.WithFields(logrus.Fields{
			"event":   "FindSubsCharge",
			"orderId": noti.OutTradeNo,
		}).Error(err)

		if err == sql.ErrNoRows {
			if _, err := w.Write([]byte(success)); err != nil {
				logger.Error(err)
			}
		}
		if _, err := w.Write([]byte(fail)); err != nil {
			logger.Error(err)
		}
		return
	}

	// 2、判断total_amount是否确实为该订单的实际金额（即商户订单创建时的金额）
	if charge.AliPrice() != noti.TotalAmount {
		logger.WithFields(logrus.Fields{
			"event":   "PaymentFailed",
			"orderId": noti.OutTradeNo,
		}).Infof("Expected net price: %s, actually received: %s", charge.AliPrice(), noti.TotalAmount)

		if _, err := w.Write([]byte(success)); err != nil {
			logger.Error(err)
		}

		return
	}

	// If this order already confirmed.
	if charge.IsConfirmed {
		logger.WithFields(logrus.Fields{
			"event":   "AlreadyConfirmed",
			"orderId": noti.OutTradeNo,
		}).Info("Duplicate notification since this order is already confirmed.")

		if _, err := w.Write([]byte(success)); err != nil {
			logger.Error(err)
		}
		return
	}

	confirmedSubs, result := router.model.ConfirmPayment(orderID, util.ParseAliTime(noti.GmtPayment))

	if result != nil {
		logger.WithFields(logrus.Fields{
			"event":   "ConfirmOrder",
			"orderId": noti.OutTradeNo,
		}).Error(err)

		go func() {
			err := router.model.SaveConfirmationResult(result)
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

	logger.WithFields(logrus.Fields{
		"event":   "PaymentFailed",
		"OrderId": noti.OutTradeNo,
	}).Infof("Confirmed at %s, membership from %s to %s", confirmedSubs.ConfirmedAt, confirmedSubs.StartDate, confirmedSubs.EndDate)

	go func() {
		if err := router.sendConfirmationEmail(confirmedSubs); err != nil {
			logger.Error(err)
		}
	}()

	go func() {
		if err := router.model.SaveConfirmationResult(paywall.NewConfirmationSucceeded(orderID)); err != nil {
			logger.Error(err)
		}
	}()

	if b, err := w.Write([]byte(success)); err != nil {
		logger.WithField("byte_output", b).Error(err)
	}
}
