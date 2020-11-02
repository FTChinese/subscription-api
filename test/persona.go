// +build !production

package test

import (
	"time"

	"github.com/FTChinese/go-rest/chrono"
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

	Orders map[string]subs.Order
	Member reader.Membership
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

		Orders: make(map[string]subs.Order),
		Member: reader.Membership{},
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
	switch p.kind {
	case enum.AccountKindFtc:
		return reader.FtcAccount{
			FtcID:    p.FtcID,
			UnionID:  null.String{},
			StripeID: null.StringFrom(p.StripeID),
			Email:    p.Email,
			UserName: null.StringFrom(p.UserName),
		}

	case enum.AccountKindWx:
		return reader.FtcAccount{
			FtcID:    "",
			UnionID:  null.StringFrom(p.UnionID),
			StripeID: null.String{},
			Email:    "",
			UserName: null.String{},
		}

	case enum.AccountKindLinked:
		return reader.FtcAccount{
			FtcID:    p.FtcID,
			UnionID:  null.StringFrom(p.UnionID),
			StripeID: null.StringFrom(p.StripeID),
			Email:    p.Email,
			UserName: null.StringFrom(p.UserName),
		}
	}

	return reader.FtcAccount{}
}

func (p *Persona) Membership() reader.Membership {
	m := reader.Membership{
		MemberID:      p.AccountID(),
		Edition:       p.plan.Edition,
		ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
		PaymentMethod: p.payMethod,
		FtcPlanID:     null.String{},
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        enum.SubsStatusNull,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
	}

	if p.expired {
		m.ExpireDate = chrono.DateFrom(time.Now().AddDate(0, -6, 0))
	}

	switch p.payMethod {
	case enum.PayMethodWx, enum.PayMethodAli:
		m.FtcPlanID = null.StringFrom(p.plan.ID)

	case enum.PayMethodStripe:
		m.StripeSubsID = null.StringFrom(faker.GenStripeSubID())
		m.StripePlanID = null.StringFrom(faker.GenStripePlanID())
		m.AutoRenewal = true
		m.Status = enum.SubsStatusActive

	case enum.PayMethodApple:
		m.AppleSubsID = null.StringFrom(p.AppleSubID)
		m.AutoRenewal = true

	case enum.PayMethodB2B:
		m.B2BLicenceID = null.StringFrom(faker.GenLicenceID())
	}

	return m.Normalize()
}

func (p *Persona) WxOrderBuilder() subs.PaymentConfig {
	return subs.NewPayment(p.FtcAccount(), p.plan).
		WithWxpay(WxPayApp)
}

func (p *Persona) AliOrderBuilder() subs.PaymentConfig {
	return subs.NewPayment(p.FtcAccount(), p.plan).
		WithAlipay()
}

func (p *Persona) CreateOrder() subs.Order {
	var payConfig subs.PaymentConfig
	if p.payMethod == enum.PayMethodWx {
		payConfig = p.WxOrderBuilder()
	} else if p.payMethod == enum.PayMethodAli {
		payConfig = p.AliOrderBuilder()
	} else {
		panic("only alipay or wxpay supported")
	}

	kind, ve := p.Member.AliWxSubsKind(p.plan.Edition)
	if ve != nil {
		panic(ve)
	}

	var bs []subs.BalanceSource
	if kind == enum.OrderKindUpgrade {
		bs = p.findBalanceSources(time.Now())
	}

	pi, err := payConfig.BuildIntent(bs, kind)
	if err != nil {
		panic(err)
	}

	p.Orders[pi.Order.ID] = pi.Order

	return pi.Order
}

func (p *Persona) ConfirmOrder(o subs.Order) subs.ConfirmationResult {

	res, err := o.Confirm(subs.PaymentResult{
		ConfirmedAt: chrono.TimeNow(),
	}, p.Member)
	if err != nil {
		panic(err)
	}

	p.Orders[res.Order.ID] = res.Order
	p.Member = res.Membership

	return res
}

func (p *Persona) findBalanceSources(anchor time.Time) []subs.BalanceSource {

	sources := make([]subs.BalanceSource, 0)

	for _, v := range p.Orders {
		if v.IsZero() || !v.IsConfirmed() {
			continue
		}

		if v.Tier != enum.TierStandard {
			continue
		}

		if v.EndDate.Time.Before(anchor) {
			continue
		}

		bs := subs.BalanceSource{
			OrderID:   v.ID,
			Amount:    v.Amount,
			StartDate: v.StartDate,
			EndDate:   v.EndDate,
		}

		sources = append(sources, bs)
	}

	return sources
}

func (p *Persona) RenewN(n int) []subs.Order {
	orders := make([]subs.Order, 0)

	for i := 0; i < n; i++ {
		o := p.CreateOrder()
		p.ConfirmOrder(o)

		orders = append(orders, o)
	}

	return orders
}

func (p *Persona) PaymentResult(order subs.Order) subs.PaymentResult {

	switch p.payMethod {
	case enum.PayMethodWx:
		result := subs.NewWxWebhookResult(NewWxWHUnsigned(order))
		return result

	case enum.PayMethodAli:
		n := AliNoti(order)
		result, err := subs.NewAliWebhookResult(&n)
		if err != nil {
			panic(err)
		}
		return result

	default:
		panic("Not ali or wx pay")
	}
}

func GenProratedOrders(upOrderID string) []subs.ProratedOrder {
	orders := make([]subs.ProratedOrder, 0)

	orders = append(orders, subs.ProratedOrder{
		OrderID:        subs.MustGenerateOrderID(),
		Balance:        99,
		CreatedUTC:     chrono.TimeNow(),
		ConsumedUTC:    chrono.Time{},
		UpgradeOrderID: upOrderID,
	})

	return orders
}
