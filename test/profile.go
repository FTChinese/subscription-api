package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/icrowley/fake"
	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/wxlogin"
	"os"
	"time"
)

type Profile struct {
	FtcID      string
	UnionID    string
	Email      string
	Password   string
	UserName   string
	Avatar     string
	OpenID     string
	ExpireDate chrono.Date
	IP         string
}

func NewProfile() Profile {
	profile := randomdata.GenerateProfile(randomdata.RandomGender)
	return Profile{
		FtcID:      uuid.New().String(),
		UnionID:    GenWxID(),
		Email:      profile.Email,
		Password:   profile.Login.Password,
		UserName:   profile.Login.Username,
		Avatar:     profile.Picture.Thumbnail,
		OpenID:     GenWxID(),
		ExpireDate: chrono.DateNow(),
		IP:         randomdata.IpV4Address(),
	}
}

var MyProfile = Profile{
	FtcID:    MyFtcID,
	UnionID:  MyUnionID,
	Email:    MyEmail,
	Password: "12345678",
	UserName: "weiguo.ni",
	Avatar:   "http://thirdwx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTIibCfVIicoNXZ15Af6nWkXwq5QgFcrNdkEKMHT7P1oJVI6McLT2qFia2ialF4FSMnm33yS0eAq7MK1cA/132",
}

func (p Profile) User(id ID) paywall.User {

	var user paywall.User

	switch id {
	case IDFtc:
		user, _ = paywall.NewUser(null.StringFrom(p.FtcID), null.String{})

	case IDWx:
		user, _ = paywall.NewUser(null.String{}, null.StringFrom(p.UnionID))

	case IDBound:
		user, _ = paywall.NewUser(null.StringFrom(p.FtcID), null.StringFrom(p.UnionID))
	}

	return user
}

func (p Profile) RandomUser() paywall.User {
	return p.User(ID(randomdata.Number(0, 3)))
}

func (p Profile) FtcUser() paywall.FtcUser {
	return paywall.FtcUser{
		UserID:   p.FtcID,
		Email:    p.Email,
		UserName: null.StringFrom(p.UserName),
	}
}

// BuildSubs generates Subscription for the following
// combination matrix:
// ftcOnlyId       wechatPay   create
// wechatOnlyId    aliPay      renew
// boundId					   upgrade
func (p Profile) BuildSubs(u paywall.User, pm enum.PayMethod, k paywall.SubsKind) paywall.Subscription {

	var subs paywall.Subscription
	var err error

	if k == paywall.SubsKindUpgrade {
		subs, err = paywall.NewSubsUpgrade(
			u,
			GenUpgradePlan())

		if err != nil {
			panic(err)
		}
	} else {
		subs, err = paywall.NewSubs(
			u,
			YearlyStandard)

		if err != nil {
			panic(err)
		}

		subs.Kind = k
	}

	switch pm {
	case enum.PayMethodWx:
		subs = subs.WithWxpay(WxPayClient.GetApp().AppID)
	case enum.PayMethodAli:
		subs = subs.WithAlipay()
	}

	return subs
}

// SubsRandom builds a random subscription order.
func (p Profile) SubsRandom(u paywall.User) paywall.Subscription {
	return p.BuildSubs(
		p.RandomUser(),
		enum.PayMethod(randomdata.Number(1, 3)),
		paywall.SubsKind(randomdata.Number(1, 3)),
	)
}

func (p Profile) SubsCreate(u paywall.User) paywall.Subscription {
	return p.BuildSubs(
		u,
		enum.PayMethod(randomdata.Number(1, 3)),
		paywall.SubsKindCreate,
	)
}

func (p Profile) SubsRenew(u paywall.User) paywall.Subscription {
	return p.BuildSubs(
		u,
		enum.PayMethod(randomdata.Number(1, 3)),
		paywall.SubsKindRenew,
	)
}

func (p Profile) SubsUpgrade(u paywall.User) paywall.Subscription {
	return p.BuildSubs(
		u,
		enum.PayMethod(randomdata.Number(1, 3)),
		paywall.SubsKindUpgrade,
	)
}

func (p Profile) SubsConfirmed(u paywall.User) paywall.Subscription {
	subs := p.SubsRandom(u)

	subs, err := subs.ConfirmWithMember(paywall.Membership{}, time.Now())

	if err != nil {
		panic(err)
	}

	return subs
}

func (p Profile) WxAccess() wxlogin.OAuthAccess {
	acc := wxlogin.OAuthAccess{
		AccessToken:  GenToken(),
		ExpiresIn:    7200,
		RefreshToken: GenToken(),
		OpenID:       p.OpenID,
		Scope:        "snsapi_userinfo",
		UnionID:      null.StringFrom(p.UnionID),
	}
	acc.GenerateSessionID()
	acc.CreatedAt = chrono.TimeNow()
	acc.UpdatedAt = chrono.TimeNow()
	return acc
}

func (p Profile) WxInfo() wxlogin.UserInfo {
	return wxlogin.UserInfo{
		UnionID:    p.UnionID,
		NickName:   randomdata.SillyName(),
		AvatarURL:  p.Avatar,
		Sex:        randomdata.Number(0, 3),
		Country:    randomdata.Country(randomdata.FullCountry),
		Province:   randomdata.ProvinceForCountry("US"),
		City:       randomdata.City(),
		Privileges: []string{},
	}
}

func AliNoti() alipay.TradeNotification {
	return alipay.TradeNotification{
		NotifyTime: time.Now().In(time.UTC).Format(chrono.SQLDateTime),
		NotifyType: "trade_status_sync",
		NotifyId:   fake.CharactersN(36),
		AppId:      os.Getenv("ALIPAY_APP_ID"),
		Charset:    "utf-8",
		Version:    "1.0",
		SignType:   "RSA2",
		Sign:       fake.CharactersN(256),
		TradeNo:    fake.CharactersN(64),
		OutTradeNo: fake.CharactersN(18),
		GmtCreate:  time.Now().In(time.UTC).Format(chrono.SQLDateTime),
		GmtPayment: time.Now().In(time.UTC).Format(chrono.SQLDateTime),
	}
}
