package test

import (
	"fmt"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/pkg/client"
	"github.com/FTChinese/subscription-api/pkg/redeem"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/Pallinder/go-randomdata"
	"github.com/brianvoe/gofakeit/v4"
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"github.com/smartwalle/alipay"
	"log"
	"os"
	"strings"
	"time"
)

func RandomClientApp() client.Client {
	return client.Client{
		ClientType: enum.Platform(randomdata.Number(1, 4)),
		Version:    null.StringFrom(GenVersion()),
		UserIP:     null.StringFrom(randomdata.IpV4Address()),
		UserAgent:  null.StringFrom(randomdata.UserAgentString()),
	}
}

// GenVersion creates a semantic version string.
func GenVersion() string {
	return fmt.Sprintf("%d.%d.%d", randomdata.Number(10), randomdata.Number(1, 10), randomdata.Number(1, 10))
}

func MustGenOrderID() string {
	id, err := subs.GenerateOrderID()
	if err != nil {
		panic(err)
	}

	return id
}

func GetCusID() string {
	id, _ := gorest.RandomBase64(9)
	return "cus_" + id
}

func GenWxID() string {
	id, _ := gorest.RandomBase64(21)
	return id
}

func GenToken() string {
	token, _ := gorest.RandomBase64(82)
	return token
}

func RandomPayMethod() enum.PayMethod {
	return enum.PayMethod(randomdata.Number(1, 3))
}

func GenAvatar() string {
	var gender = []string{"men", "women"}

	n := randomdata.Number(1, 35)
	g := gender[randomdata.Number(0, 2)]

	return fmt.Sprintf("https://randomuser.me/api/portraits/thumb/%s/%d.jpg", g, n)
}

const charset = "0123456789"

func randNumericString() string {
	return rand.StringWithCharset(9, charset)
}

func GenAppleSubID() string {
	return "1000000" + randNumericString()
}

func SimplePassword() string {
	return gofakeit.Password(true, false, true, false, false, 8)
}
func WxXMLNotification(orderID string) string {
	openID, _ := gorest.RandomBase64(21)
	nonce, _ := gorest.RandomHex(16)

	noti := wechat.Notification{
		OpenID:        null.StringFrom(openID),
		IsSubscribed:  false,
		TradeType:     null.StringFrom("APP"),
		BankType:      null.StringFrom("CMC"),
		TotalFee:      null.IntFrom(25800),
		TransactionID: null.StringFrom(rand.String(28)),
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

func WxNotification(orderID string) wechat.Notification {
	n := WxXMLNotification(orderID)

	p, err := wechat.DecodeXML(strings.NewReader(n))

	if err != nil {
		panic(err)
	}

	return wechat.NewNotification(p)
}

func WxXMLPrepay() string {
	nonce, _ := gorest.RandomHex(16)

	uni := wechat.UnifiedOrderResp{
		PrepayID: null.StringFrom(rand.String(36)),
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

func WxPrepay(orderID string) wechat.UnifiedOrderResp {
	uni := WxXMLPrepay()

	p, err := wechat.DecodeXML(strings.NewReader(uni))

	if err != nil {
		log.Fatal(err)
	}

	return wechat.NewUnifiedOrderResp(orderID, p)
}

func AliNoti() alipay.TradeNotification {
	return alipay.TradeNotification{
		NotifyTime: time.Now().In(time.UTC).Format(chrono.SQLDateTime),
		NotifyType: "trade_status_sync",
		NotifyId:   rand.String(36),
		AppId:      os.Getenv("ALIPAY_APP_ID"),
		Charset:    "utf-8",
		Version:    "1.0",
		SignType:   "RSA2",
		Sign:       rand.String(256),
		TradeNo:    rand.String(64),
		OutTradeNo: rand.String(18),
		GmtCreate:  time.Now().In(time.UTC).Format(chrono.SQLDateTime),
		GmtPayment: time.Now().In(time.UTC).Format(chrono.SQLDateTime),
	}
}

func GenCardSerial() string {
	now := time.Now()
	anni := now.Year() - 2005
	suffix := randomdata.Number(0, 9999)

	return fmt.Sprintf("%d%02d%04d", anni, now.Month(), suffix)
}

func GiftCard() redeem.GiftCard {
	code, _ := gorest.RandomHex(8)

	return redeem.GiftCard{
		Code:       strings.ToUpper(code),
		Tier:       enum.TierStandard,
		CycleUnit:  enum.CycleYear,
		CycleValue: null.IntFrom(1),
	}
}
