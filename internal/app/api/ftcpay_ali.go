package api

import (
	"net/http"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
)

// AliPay creates an http handler function depending
// on the device platform.
//
//	POST /<desktop|mobile|app>/{tier}/{cycle}?<test=true>&<return_url=xxx>
//
// Header: X-User-Id or X-Union-Id, or both.
// `return_url` parameter is only required for apps running on ftacademy.cn. It is deprecated and changed to request body.
//
// Input:
// priceId: string;
// discountId?: string;
// returnUrl?: string; Only for browsers.
func (routes FtcPayRoutes) AliPay(kind ali.EntryKind) http.HandlerFunc {
	webhookURL := config.AliWxWebhookURL(
		routes.live,
		enum.PayMethodAli)

	return func(w http.ResponseWriter, req *http.Request) {
		defer routes.Logger.Sync()
		sugar := routes.Logger.Sugar()

		// Collect client metadata from header.
		clientApp := footprint.NewClient(req)
		// Get user compound ids from header.
		readerIDs := ids.UserIDsFromHeader(req.Header)

		sugar.Infof("Alipay from app for %v", readerIDs)

		// Find user account.
		acnt, err := routes.ReaderRepo.FindBaseAccount(readerIDs)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).DBError(err)
			return
		}

		sugar.Infof("Base account found %v", acnt)

		var params ftcpay.AliPayReq
		if err := gorest.ParseJSON(req.Body, &params); err != nil {
			sugar.Error(err)
			_ = render.New(w).BadRequest(err.Error())
			return
		}

		sugar.Infof("Request body parsed %v", params)

		// TODO: ensure return url is set.
		if ve := params.Validate(); ve != nil {
			sugar.Error(err)
			_ = render.New(w).Unprocessable(ve)
			return
		}

		sugar.Infof("Start loading checkout item...")

		item, re := routes.loadCheckoutItem(params.FtcCartParams, routes.live)
		if re != nil {
			sugar.Error(re)
			_ = render.New(w).JSON(re.StatusCode, re)
			return
		}

		sugar.Infof("Checkout item loaded %v", item)

		cart := reader.NewShoppingCart(acnt).
			WithFtcItem(item).
			WithAlipay()

		sugar.Infof("Creating order...")
		pi, err := routes.SubsRepo.CreateOrder(cart)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).InternalServerError(err.Error())
			return
		}

		sugar.Infof("Created order: %+v", pi.Order)

		err = routes.postOrderCreation(pi.Order, clientApp)
		if err != nil {
			_ = render.New(w).DBError(err)
			return
		}

		or := ali.OrderReq{
			Title:       pi.Order.PaymentTitle(),
			FtcOrderID:  pi.Order.ID,
			TotalAmount: pi.Order.AliPayable(),
			WebhookURL:  webhookURL,
			TxKind:      kind,
			ReturnURL:   params.ReturnURL,
		}

		param, err := routes.AliPayClient.CreateOrder(or)
		if err != nil {
			sugar.Error(err)
			_ = xhttp.HandleSubsErr(w, err)
			return
		}
		sugar.Infof("Alipay signed order param: %s", param)

		alipayIntent, err := ftcpay.NewAliPaymentIntent(
			pi,
			param,
			kind)
		if err != nil {
			_ = render.New(w).InternalServerError(err.Error())
			return
		}

		go func() {
			err := routes.SubsRepo.SavePaymentIntent(alipayIntent.Schema())
			if err != nil {
				sugar.Error(err)
			}
		}()

		_ = render.New(w).OK(alipayIntent)
	}
}

// AliWebHook handles alipay server-side notification.
// See https://opendocs.alipay.com/open/204/105301
// POST /webhook/alipay
func (routes FtcPayRoutes) AliWebHook(w http.ResponseWriter, req *http.Request) {
	defer routes.Logger.Sync()
	sugar := routes.Logger.Sugar()

	var send = func(ok bool) {
		var err error
		if !ok {
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

	// Use this to test mocking data since we don't know how
	// signature is generated for ali webhook.
	//payload := test.GetMockPayload(req)
	//If err is nil, then the signature is verified.
	payload, err := routes.AliPayClient.GetWebhookPayload(req)
	sugar.Infof("+%v", payload)
	if err != nil {
		sugar.Error(err)
		send(false)
		return
	}

	go func() {
		sugar.Infof("Saving alipay webhook payload...")
		err := routes.SubsRepo.SaveAliWebhookPayload(
			ali.NewWebhookPayload(payload))
		if err != nil {
			sugar.Error(err)
		}
	}()

	sugar.Info("Start processing ali webhook")
	payResult, err := ftcpay.NewAliWebhookResult(payload)

	// 1、商户需要验证该通知数据中的out_trade_no是否为商户系统中创建的订单号
	// 2、判断total_amount是否确实为该订单的实际金额（即商户订单创建时的金额）
	_, cfmErr := routes.processWebhookResult(payResult)

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
