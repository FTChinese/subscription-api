// +build !production

package test

import (
	"github.com/FTChinese/subscription-api/pkg/addon"
	price2 "github.com/FTChinese/subscription-api/pkg/price"
	"time"

	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/google/uuid"
	"github.com/guregu/null"
)

// Persona mocks a user.
type Persona struct {
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

	kind      enum.AccountKind
	price     price2.FtcPrice
	payMethod enum.PayMethod
	expired   bool
	autoRenew bool
	reserved  addon.AddOn
}

func NewPersona() *Persona {
	faker.SeedGoFake()

	return &Persona{
		FtcID:      uuid.New().String(),
		UnionID:    faker.GenWxID(),
		StripeID:   faker.GenCustomerID(),
		Email:      gofakeit.Email(),
		Password:   faker.SimplePassword(),
		UserName:   gofakeit.Username(),
		Nickname:   gofakeit.Name(),
		Avatar:     faker.GenAvatar(),
		OpenID:     faker.GenWxID(),
		IP:         gofakeit.IPv4Address(),
		AppleSubID: faker.GenAppleSubID(),

		kind:      enum.AccountKindFtc,
		price:     faker.PriceStdYear,
		payMethod: enum.PayMethodAli,
		expired:   false,
		autoRenew: false,
	}
}

func (p *Persona) SetAccountKind(k enum.AccountKind) *Persona {
	p.kind = k
	return p
}

func (p *Persona) SetPrice(pp price2.FtcPrice) *Persona {
	p.price = pp
	return p
}

func (p *Persona) SetReservedDays(r addon.AddOn) *Persona {
	p.reserved = r

	return p
}

func (p *Persona) GetPlan() price2.FtcPrice {
	return p.price
}

func (p *Persona) SetPayMethod(m enum.PayMethod) *Persona {
	p.payMethod = m
	return p
}

func (p *Persona) SetExpired(expired bool) *Persona {
	p.expired = expired
	return p
}

func (p *Persona) SetAutoRenew(t bool) *Persona {
	p.autoRenew = t
	return p
}

func (p *Persona) AccountID() reader.MemberID {

	var id reader.MemberID

	switch p.kind {
	case enum.AccountKindFtc:
		id = reader.MemberID{
			CompoundID: p.FtcID,
			FtcID:      null.StringFrom(p.FtcID),
			UnionID:    null.String{},
		}

	case enum.AccountKindWx:
		id = reader.MemberID{
			CompoundID: p.UnionID,
			FtcID:      null.String{},
			UnionID:    null.StringFrom(p.UnionID),
		}

	case enum.AccountKindLinked:
		id = reader.MemberID{
			CompoundID: p.FtcID,
			FtcID:      null.StringFrom(p.FtcID),
			UnionID:    null.StringFrom(p.UnionID),
		}
	}

	return id
}

func (p *Persona) FtcAccount() reader.FtcAccount {
	switch p.kind {
	case enum.AccountKindFtc:
		return reader.FtcAccount{
			FtcID:    p.FtcID,
			UnionID:  null.String{},
			StripeID: null.String{},
			Email:    p.Email,
			UserName: null.StringFrom(p.UserName),
			VIP:      false,
		}

	case enum.AccountKindWx:
		return reader.FtcAccount{
			FtcID:    "",
			UnionID:  null.StringFrom(p.UnionID),
			StripeID: null.String{},
			Email:    p.Email,
			UserName: null.StringFrom(p.UserName),
			VIP:      false,
		}

	case enum.AccountKindLinked:
		return reader.FtcAccount{
			FtcID:    p.FtcID,
			UnionID:  null.StringFrom(p.UnionID),
			StripeID: null.String{},
			Email:    p.Email,
			UserName: null.StringFrom(p.UserName),
			VIP:      false,
		}
	}

	return reader.FtcAccount{}
}

func (p *Persona) Membership() reader.Membership {
	expiresDate := time.Now().AddDate(1, 0, 1)
	if p.expired {
		expiresDate = time.Now().AddDate(0, -6, 0)
	}

	return reader.NewMockMemberBuilder("").
		WithIDs(p.AccountID()).
		WithExpiration(expiresDate).
		WithPayMethod(p.payMethod).
		WithPrice(p.price.Price).
		WithAddOn(p.reserved).
		WithAutoRenewal(p.autoRenew).
		Build()
}

func (p *Persona) NewOrder(k enum.OrderKind) subs.Order {
	return subs.NewMockOrderBuilder("").
		WithUserIDs(p.AccountID()).
		WithPrice(p.price).
		WithKind(k).
		WithPayMethod(p.payMethod).
		Build()
}
