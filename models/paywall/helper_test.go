package paywall

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

type AccountKind int

const (
	AccountKindFtc AccountKind = iota
	AccountKindWx
	AccountKindLinked
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
	return Profile{
		FtcID:      uuid.New().String(),
		UnionID:    GenWxID(),
		Email:      fake.EmailAddress(),
		Password:   fake.SimplePassword(),
		UserName:   fake.UserName(),
		Avatar:     GenAvatar(),
		OpenID:     GenWxID(),
		ExpireDate: chrono.DateNow(),
		IP:         fake.IPv4(),
	}
}

func (p Profile) UserID(kind AccountKind) reader.AccountID {
	var id reader.AccountID

	switch kind {
	case AccountKindFtc:
		id, _ = reader.NewID(p.FtcID, "")

	case AccountKindWx:
		id, _ = reader.NewID("", p.UnionID)

	case AccountKindLinked:
		id, _ = reader.NewID(p.FtcID, p.UnionID)
	}

	return id
}

func (p Profile) RandomKindUserID() reader.AccountID {
	return p.UserID(AccountKind(randomdata.Number(0, 3)))
}

func (p Profile) FtcUser() reader.Account {
	return reader.Account{
		FtcID:    p.FtcID,
		UnionID:  null.String{},
		StripeID: null.String{},
		Email:    p.Email,
		UserName: null.StringFrom(p.UserName),
	}
}

func (p Profile) Membership(kind AccountKind, pm enum.PayMethod, expired bool) Membership {
	m := NewMember(p.UserID(kind))
	m.Tier = standardYearlyPlan.Tier
	m.Cycle = standardYearlyPlan.Cycle

	if expired {
		m.ExpireDate = chrono.DateFrom(time.Now().AddDate(0, 0, -7))
	} else {
		m.ExpireDate = chrono.DateFrom(time.Now().AddDate(1, 0, 1))
	}

	m.PaymentMethod = pm

	if pm == enum.PayMethodStripe {
		m.StripeSubID = null.StringFrom(GenSubID())
		m.StripePlanID = null.StringFrom(stripePlanIDsTest["standard_year"])
		m.AutoRenewal = true
		m.Status = SubStatusActive
	}

	return m
}

func GetWxAppID() string {
	return viper.GetString("wxapp.m_subs.app_id")
}
