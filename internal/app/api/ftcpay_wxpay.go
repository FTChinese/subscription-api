package api

import (
	"net/http"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"github.com/objcoding/wxpay"
)

// WxPay creates order for wechat pay.
// Header: X-User-Id: <uuid>, or X-Union-Id: <wechat union id>, or both.
// Input:
// priceId: string;
// discountId?: string;
// openId?: string; Required only for payment inside wechat in-house browser.
func (routes FtcPayRoutes) WxPay(tradeType wechat.TradeType) http.HandlerFunc {

	webhookURL := config.AliWxWebhookURL(
		routes.live,
		enum.PayMethodWx)

	// Find the client to use for wxpay
	payClient, err := routes.WxPayClients.FindByPlatform(tradeType)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		defer routes.Logger.Sync()
		sugar := routes.Logger.Sugar()

		sugar.Info("Start creating a wechat order")

		clientMeta := footprint.NewClient(req)
		readerIDs := ids.UserIDsFromHeader(req.Header)

		// Find user account.
		acnt, err := routes.ReaderRepo.FindBaseAccount(readerIDs)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).DBError(err)
			return
		}

		// Parse request body.
		input := ftcpay.NewWxPayReq(tradeType)
		if err := gorest.ParseJSON(req.Body, &input); err != nil {
			sugar.Error(err)
			_ = render.New(w).BadRequest(err.Error())
			return
		}
		if ve := input.Validate(); ve != nil {
			sugar.Error(err)
			_ = render.New(w).Unprocessable(ve)
			return
		}

		item, re := routes.loadCheckoutItem(input.FtcCartParams, routes.live)
		if re != nil {
			sugar.Error(re)
			_ = render.New(w).JSON(re.StatusCode, re)
			return
		}

		cart := reader.NewShoppingCart(acnt).
			WithFtcItem(item).
			WithWxPay(payClient.GetApp().AppID)

		pi, err := routes.SubsRepo.CreateOrder(cart)
		if err != nil {
			sugar.Error(err)
			_ = xhttp.HandleSubsErr(w, err)
			return
		}

		sugar.Infof("Created order: %+v", pi.Order)

		err = routes.postOrderCreation(pi.Order, clientMeta)
		if err != nil {
			_ = render.New(w).DBError(err)
			return
		}

		// 商户后台收到用户支付单，调用微信支付统一下单接口
		// Native app https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=8_3
		// QR: https://pay.weixin.qq.com/wiki/doc/api/native.php?chapter=9_1
		// UnifiedOrder checks if `return_code` is SUCCESS/FAIL,
		// validate the signature
		// You have to check if return_code == SUCCESS, appid, mch_id, result_code are valid.
		orderReq := wechat.NewOrderReq(wechat.UnifiedOrderParams{
			Body:        pi.Order.PaymentTitle(),
			OutTradeNo:  pi.Order.ID,
			TotalAmount: pi.Order.WxPayable(),
			UserIP:      clientMeta.UserIP.String,
			WebhookURL:  webhookURL,
			TradeType:   string(tradeType),
			OpenID:      input.OpenID,
		})

		orderPayload, err := payClient.CreateOrder(orderReq)

		// Save raw response.
		go func() {
			err := routes.SubsRepo.SaveWxPayload(
				wechat.NewPayloadSchema(
					pi.Order.ID,
					orderPayload,
				).WithKind(wechat.RowKindCreateOrder),
			)
			if err != nil {
				sugar.Error(err)
			}
		}()

		if err != nil {
			sugar.Error(err)
			_ = render.New(w).BadRequest(err.Error())
			return
		}

		err = payClient.GetApp().ValidateOrderPayload(orderPayload)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).InternalServerError(err.Error())
			return
		}

		// 统一下单接口返回正常的prepay_id，再按签名规范重新生成签名后，将数据传输给APP。
		// 参与签名的字段名为appid，partnerid，prepayid，noncestr，timestamp，package。注意：package的值格式为Sign=WXPay
		payParams, err := payClient.SDKParams(
			wechat.NewOrderResp(orderPayload),
			tradeType)
		if err != nil {
			_ = render.NewBadRequest(err.Error())
			return
		}

		payIntent := ftcpay.NewWxPaymentIntent(pi, payParams)

		go func() {
			err := routes.SubsRepo.SavePaymentIntent(payIntent.Schema())
			if err != nil {
				sugar.Error(err)
			}
		}()

		_ = render.New(w).OK(payIntent)
	}
}

// WxWebHook implements 支付结果通知
// https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_7&index=3
func (routes FtcPayRoutes) WxWebHook(w http.ResponseWriter, req *http.Request) {
	defer routes.Logger.Sync()
	sugar := routes.Logger.Sugar()

	sugar.Info("Wxpay webhook received message")

	resp := wxpay.Notifies{}

	var send = func(err error) {
		var e error
		if err != nil {
			_, e = w.Write([]byte(resp.NotOK(err.Error())))
		} else {
			_, e = w.Write([]byte(resp.OK()))
		}

		if e != nil {
			sugar.Error(e)
		}
	}

	defer req.Body.Close()

	// Decode Wechat XML request body.
	// If it cannot be decoded, tell wechat to resend it.
	rawPayload, err := wechat.DecodeXML(req.Body)
	if err != nil {
		sugar.Error(err)
		send(err)
		return
	}

	if err := wechat.ValidateWebhookPayload(rawPayload); err != nil {
		sugar.Error(err)
		send(err)
		return
	}

	// Verify signature.
	client, err := routes.WxPayClients.FindByAppID(wechat.GetAppID(rawPayload))
	if err != nil {
		sugar.Error(err)
		send(err)
		return
	}

	err = client.VerifySignature(rawPayload)
	if err != nil {
		sugar.Error(err)
		send(err)
		return
	}

	go func() {
		sugar.Info("Saving wxpay webhook raw payload")
		err := routes.SubsRepo.SaveWxPayload(
			wechat.NewPayloadSchema(
				wechat.GetOrderID(rawPayload),
				rawPayload,
			).WithKind(wechat.RowKindWebhook),
		)
		if err != nil {
			sugar.Error(err)
		}
	}()

	payResult := ftcpay.NewWxWebhookResult(wechat.NewWebhookParams(rawPayload))

	sugar.Info("Start processing wx webhook")
	_, cfmErr := routes.processWebhookResult(payResult)

	// Handle confirmation error.
	if cfmErr != nil {
		sugar.Error(cfmErr)
		if cfmErr.Retry {
			send(cfmErr)
		} else {
			send(nil)
		}

		return
	}

	send(nil)
}
