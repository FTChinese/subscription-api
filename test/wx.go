// +build !production

package test

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
)

func NewWxWHUnsigned(order subs.Order) wechat.Notification {
	nonce, _ := gorest.RandomHex(16)

	noti := wechat.Notification{
		OpenID:        null.StringFrom(faker.GenWxID()),
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

	return noti
}

func NewWxOrderUnsigned() wechat.OrderResp {
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

	return or
}

func (p *Persona) WxAccess() wxlogin.OAuthAccess {
	acc := wxlogin.OAuthAccess{
		AccessToken:  faker.GenWxAccessToken(),
		ExpiresIn:    7200,
		RefreshToken: faker.GenWxAccessToken(),
		OpenID:       p.OpenID,
		Scope:        "snsapi_userinfo",
		UnionID:      null.StringFrom(p.UnionID),
	}
	acc.GenerateSessionID()
	acc.CreatedAt = chrono.TimeNow()
	acc.UpdatedAt = chrono.TimeNow()
	return acc
}

func (p *Persona) WxUser() wxlogin.UserInfoSchema {
	faker.SeedGoFake()
	return wxlogin.UserInfoSchema{
		UserInfoShared: wxlogin.UserInfoShared{
			UnionID:   p.UnionID,
			OpenID:    "",
			NickName:  null.StringFrom(gofakeit.Username()),
			AvatarURL: null.StringFrom(faker.GenAvatar()),
			Country:   null.StringFrom(gofakeit.Country()),
			Province:  null.StringFrom(gofakeit.State()),
			City:      null.StringFrom(gofakeit.City()),
		},
		Gender:    faker.RandomGender(),
		Privilege: null.String{},
	}
}
