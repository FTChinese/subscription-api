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
	FtcID    string
	UnionID  string
	Email    string
	Password string
	UserName string
	Nickname string
	Avatar   string
	OpenID   string
	IP       string
}

func NewProfile() Profile {
	return Profile{
		FtcID:    uuid.New().String(),
		UnionID:  GenWxID(),
		Email:    fake.EmailAddress(),
		Password: fake.SimplePassword(),
		UserName: fake.UserName(),
		Nickname: fake.UserName(),
		Avatar:   GenAvatar(),
		OpenID:   GenWxID(),
		IP:       fake.IPv4(),
	}
}

var MyProfile = Profile{
	FtcID:    MyFtcID,
	UnionID:  MyUnionID,
	Email:    MyEmail,
	Password: "12345678",
	UserName: "weiguo.ni",
	Nickname: fake.UserName(),
	Avatar:   "http://thirdwx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTIibCfVIicoNXZ15Af6nWkXwq5QgFcrNdkEKMHT7P1oJVI6McLT2qFia2ialF4FSMnm33yS0eAq7MK1cA/132",
	IP:       fake.IPv4(),
}

func (p Profile) UserID(kind AccountKind) paywall.UserID {

	var id paywall.UserID

	switch kind {
	case AccountKindFtc:
		id, _ = paywall.NewID(p.FtcID, "")

	case AccountKindWx:
		id, _ = paywall.NewID("", p.UnionID)

	case AccountKindLinked:
		id, _ = paywall.NewID(p.FtcID, p.UnionID)
	}

	return id
}

func (p Profile) RandomUserID() paywall.UserID {
	return p.UserID(AccountKind(randomdata.Number(0, 3)))
}

func (p Profile) FtcUser() paywall.Account {
	return paywall.Account{
		FtcID:    p.FtcID,
		UnionID:  null.String{},
		StripeID: null.String{},
		Email:    p.Email,
		UserName: null.StringFrom(p.UserName),
	}
}

func (p Profile) Membership(kind AccountKind, pm enum.PayMethod, expired bool) paywall.Membership {
	m := paywall.NewMember(p.UserID(kind))
	m.Tier = YearlyStandard.Tier
	m.Cycle = YearlyStandard.Cycle

	if expired {
		m.ExpireDate = chrono.DateFrom(time.Now().AddDate(0, 0, -7))
	} else {
		m.ExpireDate = chrono.DateFrom(time.Now().AddDate(1, 0, 1))
	}

	m.PaymentMethod = pm

	if pm == enum.PayMethodStripe {
		m.StripeSubID = null.StringFrom(GenSubID())
		m.StripePlanID = null.StringFrom(YearlyStandard.StripeID)
		m.AutoRenewal = true
		m.Status = paywall.SubStatusActive
	}

	return m
}

func (p Profile) AliWxSub(kind AccountKind, pm enum.PayMethod, usage paywall.SubsKind) paywall.Subscription {
	s, err := paywall.NewSubs(p.UserID(kind), YearlyStandard)
	if err != nil {
		panic(err)
	}

	s.ConfirmedAt = chrono.TimeNow()
	s.IsConfirmed = true
	s.EndDate = chrono.DateFrom(time.Now().AddDate(1, 0, 1))
	s.PaymentMethod = pm
	s.StartDate = chrono.DateNow()
	s.Usage = usage
	if pm == enum.PayMethodWx {
		s.WxAppID = null.StringFrom(WxPayApp.AppID)
	}

	return s
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
