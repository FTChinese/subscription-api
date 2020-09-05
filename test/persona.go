package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"time"
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

	orders map[string]subs.Order
	member reader.Membership
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

		orders: make(map[string]subs.Order),
		member: reader.Membership{},
	}
}

var MyProfile = Persona{
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

func (p *Persona) Account() reader.Account {
	switch p.kind {
	case enum.AccountKindFtc:
		return reader.Account{
			FtcID:    p.FtcID,
			UnionID:  null.String{},
			StripeID: null.StringFrom(p.StripeID),
			Email:    p.Email,
			UserName: null.StringFrom(p.UserName),
		}

	case enum.AccountKindWx:
		return reader.Account{
			FtcID:    "",
			UnionID:  null.StringFrom(p.UnionID),
			StripeID: null.String{},
			Email:    "",
			UserName: null.String{},
		}

	case enum.AccountKindLinked:
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

func (p *Persona) Membership() reader.Membership {
	m := reader.Membership{
		MemberID:      p.AccountID(),
		Edition:       p.plan.Edition,
		ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
		PaymentMethod: p.payMethod,
		FtcPlanID:     null.StringFrom(p.plan.ID),
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

func (p *Persona) CreateOrder() subs.Order {
	builder := subs.NewOrderBuilder(p.AccountID()).
		SetEnvConfig(CFG).
		SetPlan(p.plan).
		SetPayMethod(p.payMethod)

	if p.payMethod == enum.PayMethodWx {
		builder.SetWxAppID(WxPayApp.AppID).
			SetWxParams(wechat.UnifiedOrder{
				IP:        p.IP,
				TradeType: wechat.TradeTypeMobile,
				OpenID:    "",
			})
	}

	err := builder.DeduceSubsKind(p.member)
	if err != nil {
		panic(err)
	}

	if builder.GetSubsKind() == enum.OrderKindUpgrade {
		bs := p.findBalanceSources(time.Now())

		wallet := subs.NewWallet(bs, time.Now())

		builder.SetWallet(wallet)
	}

	err = builder.Build()
	if err != nil {
		panic(err)
	}

	order, _ := builder.GetOrder()

	p.orders[order.ID] = order

	return order
}

func (p *Persona) ConfirmOrder(o subs.Order) subs.ConfirmationResult {
	builder := subs.NewConfirmationBuilder(subs.PaymentResult{
		ConfirmedAt: time.Now(),
	}, false).
		SetMembership(p.member).
		SetOrder(o)

	res, err := builder.Build()
	if err != nil {
		panic(err)
	}

	p.orders[res.Order.ID] = res.Order
	p.member = res.Membership

	return res
}

func (p *Persona) findBalanceSources(anchor time.Time) []subs.BalanceSource {

	sources := make([]subs.BalanceSource, 0)

	for _, v := range p.orders {
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

func (p Persona) IAPSubs() apple.Subscription {
	s := apple.Subscription{
		Environment:           apple.EnvSandbox,
		OriginalTransactionID: p.AppleSubID,
		LastTransactionID:     faker.GenAppleSubID(),
		ProductID:             "",
		PurchaseDateUTC:       chrono.TimeNow(),
		ExpiresDateUTC:        chrono.TimeFrom(time.Now().AddDate(1, 0, 0)),
		Edition:               p.plan.Edition,
		AutoRenewal:           true,
	}

	if p.expired {
		s.ExpiresDateUTC = chrono.TimeFrom(time.Now().AddDate(-1, 0, 0))
	}

	return s
}

func (p Persona) WxAccess() wxlogin.OAuthAccess {
	acc := wxlogin.OAuthAccess{
		AccessToken:  faker.GenWxAccessTokenToken(),
		ExpiresIn:    7200,
		RefreshToken: faker.GenWxAccessTokenToken(),
		OpenID:       p.OpenID,
		Scope:        "snsapi_userinfo",
		UnionID:      null.StringFrom(p.UnionID),
	}
	acc.GenerateSessionID()
	acc.CreatedAt = chrono.TimeNow()
	acc.UpdatedAt = chrono.TimeNow()
	return acc
}

func (p Persona) WxInfo() wxlogin.UserInfo {
	faker.SeedGoFake()
	return wxlogin.UserInfo{
		UnionID:    p.UnionID,
		NickName:   gofakeit.UserAgent(),
		AvatarURL:  p.Avatar,
		Sex:        rand.IntRange(0, 3),
		Country:    gofakeit.Country(),
		Province:   gofakeit.State(),
		City:       gofakeit.City(),
		Privileges: []string{},
	}
}
