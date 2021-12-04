package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/guregu/null"
	"net/http"
)

// AliPay creates an http handler function depending
// on the device platform.
//
// 	POST /<desktop|mobile|app>/{tier}/{cycle}?<test=true>&<return_url=xxx>
//
// Header: X-User-Id or X-Union-Id, or both.
// `return_url` parameter is only required for apps running on ftacademy.cn. It is deprecated and changed to request body.
//
// Input:
// returnUrl?: string; Only for browsers.
// tier: string;
// cycle: string;
// planId?: string;
func (router SubsRouter) AliPay(kind ali.EntryKind) http.HandlerFunc {
	webhookURL := config.AliWxWebhookURL(
		router.Live,
		enum.PayMethodAli)

	return func(w http.ResponseWriter, req *http.Request) {
		defer router.Logger.Sync()
		sugar := router.Logger.Sugar()

		clientApp := footprint.NewClient(req)
		readerIDs := getReaderIDs(req.Header)

		// Find user account.
		acnt, err := router.SubsRepo.FindBaseAccount(readerIDs)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).DBError(err)
			return
		}

		var input ftcpay.AliPayReq
		if err := gorest.ParseJSON(req.Body, &input); err != nil {
			sugar.Error(err)
			_ = render.New(w).BadRequest(err.Error())
			return
		}

		// TODO: ensure return url is set.
		if ve := input.Validate(); ve != nil {
			sugar.Error(err)
			_ = render.New(w).Unprocessable(ve)
			return
		}

		item, re := router.loadCheckoutItem(input.OrderParams, !acnt.IsTest())
		if re != nil {
			sugar.Error(re)
			_ = render.New(w).JSON(re.StatusCode, re)
			return
		}

		counter := ftcpay.Counter{
			BaseAccount:  acnt,
			CheckoutItem: item,
			PayMethod:    enum.PayMethodAli,
			WxAppID:      null.String{},
		}

		pi, err := router.SubsRepo.CreateOrder(counter)
		if err != nil {
			_ = render.New(w).InternalServerError(err.Error())
			return
		}

		sugar.Infof("Created order: %+v", pi.Order)

		err = router.postOrderCreation(pi.Order, clientApp)
		if err != nil {
			_ = render.New(w).DBError(err)
			return
		}

		or := ali.OrderReq{
			Title:       subs.PaymentTitle(pi.Order.Kind, pi.Order.Edition),
			FtcOrderID:  pi.Order.ID,
			TotalAmount: pi.Order.AliPrice(),
			WebhookURL:  webhookURL,
			TxKind:      kind,
			ReturnURL:   input.ReturnURL,
		}

		param, err := router.AliPayClient.CreateOrder(or)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).InternalServerError(err.Error())
			return
		}
		sugar.Infof("Alipay signed order param: %s", param)

		switch kind {
		case ali.EntryApp:
			_ = render.New(w).OK(subs.NewAliAppPayIntent(pi, param))
			return

		case ali.EntryDesktopWeb, ali.EntryMobileWeb:
			_ = render.New(w).OK(subs.NewAliPayBrowserIntent(pi, param))
		}
	}
}

// AliWebHook handles alipay server-side notification.
// POST /webhook/alipay
func (router SubsRouter) AliWebHook(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	var send = func(ok bool) {
		var err error
		if ok {
			_, err = w.Write([]byte("fail"))
		} else {
			_, err = w.Write([]byte("success"))
		}

		if err != nil {
			sugar.Error(err)
		}
	}

	err := req.ParseForm()
	if err != nil {
		sugar.Error(err)
		send(false)
		return
	}

	// If err is nil, then the signature is verified.
	payload, err := router.AliPayClient.GetWebhookPayload(req)
	sugar.Infof("+%v", payload)
	if err != nil {
		sugar.Error(err)
		send(false)
		return
	}

	go func() {
		if err := router.SubsRepo.SaveAliNotification(*payload); err != nil {
			sugar.Error(err)
		}
	}()

	sugar.Info("Start processing ali webhook")
	payResult, err := subs.NewAliWebhookResult(payload)

	// 1、商户需要验证该通知数据中的out_trade_no是否为商户系统中创建的订单号
	// 2、判断total_amount是否确实为该订单的实际金额（即商户订单创建时的金额）
	_, cfmErr := router.processWebhookResult(payResult)

	if cfmErr != nil {
		sugar.Error(cfmErr)
		if cfmErr.Retry {
			send(false)
			return
		} else {
			send(true)
			return
		}
	}

	send(true)
}
