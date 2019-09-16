package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/icrowley/fake"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/wxlogin"
	"time"
)

type Profile struct {
	FtcID    string
	UnionID  string
	StripeID string
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
		StripeID: GetCusID(),
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
	StripeID: "cus_FOgRRgj9aMzpAv",
	Email:    MyEmail,
	Password: "12345678",
	UserName: "weiguo.ni",
	Nickname: fake.UserName(),
	Avatar:   "http://thirdwx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTIibCfVIicoNXZ15Af6nWkXwq5QgFcrNdkEKMHT7P1oJVI6McLT2qFia2ialF4FSMnm33yS0eAq7MK1cA/132",
	IP:       fake.IPv4(),
}

func (p Profile) AccountID(kind reader.AccountKind) reader.AccountID {

	var id reader.AccountID

	switch kind {
	case reader.AccountKindFtc:
		id, _ = reader.NewID(p.FtcID, "")

	case reader.AccountKindWx:
		id, _ = reader.NewID("", p.UnionID)

	case reader.AccountKindLinked:
		id, _ = reader.NewID(p.FtcID, p.UnionID)
	}

	return id
}

func (p Profile) Account(k reader.AccountKind) reader.Account {
	switch k {
	case reader.AccountKindFtc:
		return reader.Account{
			FtcID:    p.FtcID,
			UnionID:  null.String{},
			StripeID: null.StringFrom(p.StripeID),
			Email:    p.Email,
			UserName: null.StringFrom(p.UserName),
		}

	case reader.AccountKindWx:
		return reader.Account{
			FtcID:    "",
			UnionID:  null.StringFrom(p.UnionID),
			StripeID: null.String{},
			Email:    "",
			UserName: null.String{},
		}

	case reader.AccountKindLinked:
		return reader.Account{
			FtcID:    p.FtcID,
			UnionID:  null.StringFrom(p.UnionID),
			StripeID: null.StringFrom(p.StripeID),
			Email:    p.Email,
			UserName: null.StringFrom(p.UserName),
		}
	}

	return reader.Account{}
}

func (p Profile) Membership(k reader.AccountKind) paywall.Membership {
	return paywall.Membership{
		ID:           null.StringFrom(paywall.GenerateMemberID()),
		AccountID:    p.AccountID(reader.AccountKindFtc),
		LegacyTier:   null.Int{},
		LegacyExpire: null.Int{},
		Coordinate: paywall.Coordinate{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleYear,
		},
		ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
		PaymentMethod: enum.PayMethodWx,
		StripeSubID:   null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        paywall.SubStatusNull,
	}
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
