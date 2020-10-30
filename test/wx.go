package test

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"log"
	"strings"
)

func WxBaseRespToMap(r wechat.BaseResp) wxpay.Params {
	p := make(wxpay.Params)

	p.SetString("return_code", r.ReturnCode)
	p.SetString("return_msg", r.ReturnMessage)
	p.SetString("appid", r.AppID.String)
	p.SetString("mch_id", r.MID.String)
	p.SetString("nonce_str", r.Nonce.String)
	p.SetString("result_code", r.ResultCode.String)

	return p
}

func WxWebhookToMap(n wechat.Notification) wxpay.Params {
	p := WxBaseRespToMap(n.BaseResp)

	var subscribed string
	if n.IsSubscribed {
		subscribed = "Y"
	} else {
		subscribed = "N"
	}

	p.SetString("openid", n.OpenID.String)
	p.SetString("is_subscribe", subscribed)
	p.SetString("bank_type", n.BankType.String)
	p.SetInt64("total_fee", n.TotalFee.Int64)
	p.SetString("transaction_id", n.TransactionID.String)
	p.SetString("out_trade_no", n.FTCOrderID.String)
	p.SetString("time_end", n.TimeEnd.String)

	return p
}

func WxOrderRespToMap(o wechat.OrderResp) wxpay.Params {
	p := WxBaseRespToMap(o.BaseResp)

	p.SetString("prepay_id", o.PrepayID.String)

	return p
}

// WxXMLNotification mocks the data received in wechat webhook.
// To test its behavior, you must have a user row and order row in the db.
func WxXMLNotification(order subs.Order) string {
	openID, _ := gorest.RandomBase64(21)
	nonce, _ := gorest.RandomHex(16)

	noti := wechat.Notification{
		OpenID:        null.StringFrom(openID),
		IsSubscribed:  false,
		TradeType:     null.StringFrom("APP"),
		BankType:      null.StringFrom("CMC"),
		TotalFee:      null.IntFrom(order.AmountInCent()),
		TransactionID: null.StringFrom(rand.String(28)),
		FTCOrderID:    null.StringFrom(order.ID),
		TimeEnd:       null.StringFrom("20060102150405"),
	}

	noti.ReturnCode = "SUCCESS"
	noti.ReturnMessage = "OK"
	noti.AppID = null.StringFrom(WxPayApp.AppID)
	noti.MID = null.StringFrom(WxPayApp.MchID)
	noti.Nonce = null.StringFrom(nonce)
	noti.ResultCode = null.StringFrom("SUCCESS")

	p := WxWebhookToMap(noti)

	s := WxPayClient.Sign(p)

	p.SetString("sign", s)

	return wxpay.MapToXml(p)
}

func WxNotification(order subs.Order) wechat.Notification {
	n := WxXMLNotification(order)

	p, err := wechat.DecodeXML(strings.NewReader(n))

	if err != nil {
		panic(err)
	}

	return wechat.NewNotification(p)
}

// WxXMLPrepay mocks the data received from wechat as a payment intent.
func WxXMLPrepay() string {
	nonce, _ := gorest.RandomHex(16)

	or := wechat.OrderResp{
		PrepayID: null.StringFrom(rand.String(36)),
	}

	or.ReturnCode = "SUCCESS"
	or.ReturnMessage = "OK"
	or.AppID = null.StringFrom(WxPayApp.AppID)
	or.MID = null.StringFrom(WxPayApp.MchID)
	or.Nonce = null.StringFrom(nonce)
	or.ResultCode = null.StringFrom("SUCCESS")

	p := WxOrderRespToMap(or)

	s := WxPayClient.Sign(p)

	p.SetString("sign", s)

	return wxpay.MapToXml(p)
}

func WxPrepay(orderID string) wechat.OrderResp {
	uni := WxXMLPrepay()

	p, err := wechat.DecodeXML(strings.NewReader(uni))

	if err != nil {
		log.Fatal(err)
	}

	return wechat.NewOrderResp(orderID, p)
}
