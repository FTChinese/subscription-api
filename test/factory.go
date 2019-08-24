package test

import (
	"fmt"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/Pallinder/go-randomdata"
	"github.com/guregu/null"
	"github.com/icrowley/fake"
	"github.com/objcoding/wxpay"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/models/wechat"
	"log"
	"strings"
	"time"
)

func RandomClientApp() util.ClientApp {
	return util.ClientApp{
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

func GenSubID() string {
	id, _ := gorest.RandomBase64(9)
	return "sub_" + id
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

func RandomAccountKind() AccountKind {
	return AccountKind(randomdata.Number(1, 3))
}

func GenAvatar() string {
	var gender = []string{"men", "women"}

	n := randomdata.Number(1, 35)
	g := gender[randomdata.Number(0, 2)]

	return fmt.Sprintf("https://randomuser.me/api/portraits/thumb/%s/%d.jpg", g, n)
}

func GenMember(u reader.AccountID, expired bool) paywall.Membership {
	m := paywall.NewMember(u)
	m.Tier = YearlyStandard.Tier
	m.Cycle = YearlyStandard.Cycle

	if expired {
		m.ExpireDate = chrono.DateFrom(time.Now().AddDate(0, 0, -7))
	} else {
		m.ExpireDate = chrono.DateFrom(time.Now().AddDate(1, 0, 1))
	}

	return m
}

func BalanceSources() []paywall.ProrationSource {
	sources := []paywall.ProrationSource{}

	for i := 0; i < 2; i++ {
		id, _ := paywall.GenerateOrderID()
		startTime := time.Now().AddDate(i, 0, 0)
		endTime := startTime.AddDate(i+1, 0, 0)

		s := paywall.ProrationSource{
			OrderID:    id,
			PaidAmount: 258,
			StartDate:  chrono.DateFrom(startTime),
			EndDate:    chrono.DateFrom(endTime),
		}

		sources = append(sources, s)
	}

	return sources
}

func GenUpgrade(userID reader.AccountID) paywall.Upgrade {

	up := paywall.NewUpgrade(YearlyPremium).
		SetBalance(BalanceSources()).
		CalculatePayable()

	up.Member = GenMember(userID, false)

	return up
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

func WxPrepay() wechat.UnifiedOrderResp {
	uni := WxXMLPrepay()

	p, err := wechat.DecodeXML(strings.NewReader(uni))

	if err != nil {
		log.Fatal(err)
	}

	return wechat.NewUnifiedOrderResp(p)
}

func GenCardSerial() string {
	now := time.Now()
	anni := now.Year() - 2005
	suffix := randomdata.Number(0, 9999)

	return fmt.Sprintf("%d%02d%04d", anni, now.Month(), suffix)
}

func GiftCard() paywall.GiftCard {
	code, _ := gorest.RandomHex(8)

	return paywall.GiftCard{
		Code:       strings.ToUpper(code),
		Tier:       enum.TierStandard,
		CycleUnit:  enum.CycleYear,
		CycleValue: null.IntFrom(1),
	}
}
