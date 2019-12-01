package subscription

import (
	"fmt"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/icrowley/fake"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"time"
)

var gender = []string{"men", "women"}

func GenWxID() string {
	id, _ := gorest.RandomBase64(21)
	return id
}

func GenAvatar() string {
	n := randomdata.Number(1, 35)
	g := gender[randomdata.Number(0, 2)]

	return fmt.Sprintf("https://randomuser.me/api/portraits/thumb/%s/%d.jpg", g, n)
}

func GenSubID() string {
	id, _ := gorest.RandomBase64(9)
	return "sub_" + id
}

func GenCusID() string {
	id, _ := gorest.RandomBase64(9)
	return "cus_" + id
}

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
		StripeID: GenCusID(),
		Email:    fake.EmailAddress(),
		Password: fake.SimplePassword(),
		UserName: fake.UserName(),
		Nickname: fake.UserName(),
		Avatar:   GenAvatar(),
		OpenID:   GenWxID(),
		IP:       fake.IPv4(),
	}
}

func (p Profile) AccountID(kind reader.AccountKind) reader.MemberID {
	var id reader.MemberID

	switch kind {
	case reader.AccountKindFtc:
		id, _ = reader.NewMemberID(p.FtcID, "")

	case reader.AccountKindWx:
		id, _ = reader.NewMemberID("", p.UnionID)

	case reader.AccountKindLinked:
		id, _ = reader.NewMemberID(p.FtcID, p.UnionID)
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

func (p Profile) Membership(kind reader.AccountKind, pm enum.PayMethod, expired bool) Membership {
	m := NewMember(p.AccountID(kind))
	m.Tier = enum.TierStandard
	m.Cycle = enum.CycleYear

	if expired {
		m.ExpireDate = chrono.DateFrom(time.Now().AddDate(0, 0, -7))
	} else {
		m.ExpireDate = chrono.DateFrom(time.Now().AddDate(1, 0, 1))
	}

	m.PaymentMethod = pm

	if pm == enum.PayMethodStripe {
		m.StripeSubID = null.StringFrom(GenSubID())
		//m.StripePlanID = null.StringFrom(stripePlanIDsTest["standard_year"])
		m.AutoRenewal = true
		m.Status = SubStatusActive
	}

	return m
}

func GetWxAppID() string {
	return viper.GetString("wxapp.m_subs.app_id")
}
