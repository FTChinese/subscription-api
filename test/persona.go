// +build !production

package test

import (
	"github.com/FTChinese/subscription-api/pkg/pw"
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
	price     pw.FtcPrice
	payMethod enum.PayMethod
	expired   bool
	autoRenew bool
	reserved  reader.ReservedDays
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

func (p *Persona) SetPrice(pp pw.FtcPrice) *Persona {
	p.price = pp
	return p
}

func (p *Persona) SetReservedDays(r reader.ReservedDays) *Persona {
	p.reserved = r

	return p
}

func (p *Persona) GetPlan() pw.FtcPrice {
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
	return reader.MockNewFtcAccount(p.kind)
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
		WithPrice(p.price.Original).
		WithReservedDays(p.reserved).
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

func (p *Persona) AddOn() subs.AddOn {
	return subs.NewMockAddOnBuilder().
		WithUserIDs(p.AccountID()).
		WithPlan(p.price).
		BuildNew()
}

func (p *Persona) AddOnN(n int) []subs.AddOn {
	factory := subs.NewMockAddOnBuilder().
		WithUserIDs(p.AccountID()).
		WithPlan(p.price)

	var addOns []subs.AddOn
	for i := 0; i < n; i++ {
		addOns = append(addOns, factory.BuildNew())
	}

	return addOns
}
