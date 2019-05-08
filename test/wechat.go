package test

import (
	"github.com/FTChinese/go-rest"
	"github.com/guregu/null"
	"github.com/icrowley/fake"
	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/wechat"
	"log"
	"strings"
)

func MockWxXMLNotification(orderID string) string {
	openID, _ := gorest.RandomBase64(21)
	nonce, _ := gorest.RandomHex(16)

	noti := wechat.Notification{
		OpenID:        null.StringFrom(openID),
		IsSubscribed:  false,
		TradeType:     null.StringFrom("APP"),
		BankType:      null.StringFrom("CMC"),
		TotalFee:      null.IntFrom(25800),
		TransactionID: null.StringFrom(fake.CharactersN(28)),
		FTCOrderID:    null.StringFrom(orderID),
		TimeEnd:       null.StringFrom("20060102150405"),
	}

	noti.StatusCode = "SUCCESS"
	noti.StatusMessage = "OK"
	noti.AppID = null.StringFrom(WxPayApp.AppID)
	noti.MID = null.StringFrom(WxPayApp.MchID)
	noti.Nonce = null.StringFrom(nonce)
	noti.ResultCode = null.StringFrom("SUCCESS")

	p := noti.Params()

	s := WxPayClient.Sign(noti.Params())

	p.SetString("sign", s)

	return wxpay.MapToXml(p)
}

func MockWxNotification(orderID string) wechat.Notification {
	n := MockWxXMLNotification(orderID)

	p, err := wechat.DecodeXML(strings.NewReader(n))

	if err != nil {
		panic(err)
	}

	return wechat.NewNotification(p)
}

func MockWxXMLPrepay() string {
	nonce, _ := gorest.RandomHex(16)

	uni := wechat.UnifiedOrderResp{
		PrepayID: null.StringFrom(fake.CharactersN(36)),
	}

	uni.StatusCode = "SUCCESS"
	uni.StatusMessage = "OK"
	uni.AppID = null.StringFrom(WxPayApp.AppID)
	uni.MID = null.StringFrom(WxPayApp.MchID)
	uni.Nonce = null.StringFrom(nonce)
	uni.ResultCode = null.StringFrom("SUCCESS")

	p := uni.Params()

	s := WxPayClient.Sign(p)

	p.SetString("sign", s)

	return wxpay.MapToXml(p)
}

func MockWxPrepay() wechat.UnifiedOrderResp {
	uni := MockWxXMLPrepay()

	p, err := wechat.DecodeXML(strings.NewReader(uni))

	if err != nil {
		log.Fatal(err)
	}

	return wechat.NewUnifiedOrderResp(p)
}
