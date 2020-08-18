package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/models/plan"
	"github.com/FTChinese/subscription-api/models/subscription"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"github.com/Pallinder/go-randomdata"
	"github.com/brianvoe/gofakeit/v4"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"time"
)

type Profile struct {
	FtcID      string
	UnionID    string
	StripeID   string
	Email      string
	Password   string
	UserName   string
	Nickname   string
	Avatar     string
	OpenID     string
	IP         string
	AppleSubID string

	kind        reader.AccountKind
	plan        product.Plan
	expiresDate time.Time
	payMethod   enum.PayMethod
}

func NewProfile() *Profile {
	gofakeit.Seed(time.Now().UnixNano())

	return &Profile{
		FtcID:      uuid.New().String(),
		UnionID:    GenWxID(),
		StripeID:   GetCusID(),
		Email:      gofakeit.Email(),
		Password:   SimplePassword(),
		UserName:   gofakeit.Username(),
		Nickname:   gofakeit.Name(),
		Avatar:     GenAvatar(),
		OpenID:     GenWxID(),
		IP:         gofakeit.IPv4Address(),
		AppleSubID: GenAppleSubID(),

		kind:        reader.AccountKindFtc,
		plan:        YearlyStandard,
		expiresDate: time.Now().AddDate(1, 0, 0),
		payMethod:   enum.PayMethodAli,
	}
}

var MyProfile = Profile{
	FtcID:    MyFtcID,
	UnionID:  MyUnionID,
	StripeID: "cus_FOgRRgj9aMzpAv",
	Email:    MyEmail,
	Password: "12345678",
	UserName: "weiguo.ni",
	Nickname: gofakeit.Username(),
	Avatar:   "http://thirdwx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTIibCfVIicoNXZ15Af6nWkXwq5QgFcrNdkEKMHT7P1oJVI6McLT2qFia2ialF4FSMnm33yS0eAq7MK1cA/132",
	IP:       gofakeit.IPv4Address(),
}

func (p *Profile) SetAccountKind(k reader.AccountKind) *Profile {
	p.kind = k
	return p
}

func (p *Profile) SetPlan(subPlan plan.Plan) *Profile {
	p.plan = subPlan
	return p
}

func (p *Profile) SetExpireDate(t time.Time) *Profile {
	p.expiresDate = t
	return p
}

func (p *Profile) SetPayMethod(m enum.PayMethod) *Profile {
	p.payMethod = m
	return p
}

func (p Profile) AccountID() reader.MemberID {

	var id reader.MemberID

	switch p.kind {
	case reader.AccountKindFtc:
		id, _ = reader.NewMemberID(p.FtcID, "")

	case reader.AccountKindWx:
		id, _ = reader.NewMemberID("", p.UnionID)

	case reader.AccountKindLinked:
		id, _ = reader.NewMemberID(p.FtcID, p.UnionID)
	}

	return id
}

func (p Profile) Account() reader.Account {
	switch p.kind {
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

func (p Profile) Membership() subs.Membership {
	m := subs.Membership{
		MemberID:      p.AccountID(),
		Edition:       p.plan.BasePlan,
		ExpireDate:    chrono.DateFrom(p.expiresDate),
		PaymentMethod: p.payMethod,
		StripeSubID:   null.String{},
		StripePlanID:  null.String{},
		AutoRenew:     false,
		Status:        subscription.SubStatusNull,
	}

	if p.payMethod == enum.PayMethodApple {
		m.AppleSubID = null.StringFrom(p.AppleSubID)
		m.AutoRenew = true
	}

	return m
}

// StandardOrdersN generates n orders for standard membership.
func (p Profile) StandardOrdersN(n int) []subs.Order {
	orders := make([]subs.Order, 0)
	for i := 0; i < n; i++ {
		o := subs.Order{
			ID:       MustGenOrderID(),
			MemberID: p.AccountID(),
			BasePlan: p.plan.BasePlan,
			Price:    p.plan.Price,
			Charge: subs.Charge{
				Amount:   p.plan.Amount,
				Currency: p.plan.Currency,
			},
			Duration: subs.Duration{
				CycleCount: 1,
				ExtraDays:  1,
			},
			Usage:            plan.SubsKindCreate,
			PaymentMethod:    p.payMethod,
			WxAppID:          null.String{},
			UpgradeSchemaID:  null.String{},
			CreatedAt:        chrono.TimeNow(),
			ConfirmedAt:      chrono.TimeNow(),
			StartDate:        chrono.DateFrom(time.Now()),
			EndDate:          chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
			MemberSnapshotID: null.String{},
		}

		orders = append(orders, o)
	}

	return orders
}

func (p Profile) IAPSubs() apple.Subscription {
	return apple.Subscription{
		Environment:           apple.EnvSandbox,
		OriginalTransactionID: p.AppleSubID,
		LastTransactionID:     GenAppleSubID(),
		ProductID:             p.plan.AppleProductID,
		PurchaseDateUTC:       chrono.TimeNow(),
		ExpiresDateUTC:        chrono.TimeFrom(p.expiresDate),
		BasePlan:              p.plan.BasePlan,
		AutoRenewal:           true,
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
