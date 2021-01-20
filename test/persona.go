// +build !production

package test

import (
	"time"

	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/product"
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
	plan      product.ExpandedPlan
	payMethod enum.PayMethod
	expired   bool
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
		plan:      faker.PlanStdYear,
		payMethod: enum.PayMethodAli,
		expired:   false,
	}
}

func (p *Persona) SetAccountKind(k enum.AccountKind) *Persona {
	p.kind = k
	return p
}

func (p *Persona) SetPlan(subPlan product.ExpandedPlan) *Persona {
	p.plan = subPlan
	return p
}

func (p *Persona) GetPlan() product.ExpandedPlan {
	return p.plan
}

func (p *Persona) SetPayMethod(m enum.PayMethod) *Persona {
	p.payMethod = m
	return p
}

func (p *Persona) SetExpired(expired bool) *Persona {
	p.expired = expired
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
		WithPlan(p.plan.Plan).
		Build()
}

func (p *Persona) NewOrder(k enum.OrderKind) subs.Order {
	return subs.NewMockOrderBuilder("").
		WithUserIDs(p.AccountID()).
		WithPlan(p.plan).
		WithKind(k).
		WithPayMethod(p.payMethod).
		Build()
}
