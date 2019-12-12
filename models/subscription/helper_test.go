package subscription

import (
	"fmt"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/Pallinder/go-randomdata"
	"github.com/brianvoe/gofakeit/v4"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"time"
)

func init() {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

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

func GenCusID() string {
	id, _ := gorest.RandomBase64(9)
	return "cus_" + id
}

func GenAppleSubID() string {
	return "1000000" + rand.StringWithCharset(9, "0123456789")
}

func SimplePassword() string {
	return gofakeit.Password(true, false, true, false, false, 8)
}

func mustFindPlan(tier enum.Tier, cycle enum.Cycle) plan.Plan {
	p, err := plan.FindPlan(tier, cycle)
	if err != nil {
		panic(err)
	}

	return p
}

func getWxAppID() string {
	return viper.GetString("wxapp.m_subs.app_id")
}

var yearlyStandard = mustFindPlan(enum.TierStandard, enum.CycleYear)
var yearlyPremium = mustFindPlan(enum.TierPremium, enum.CycleYear)

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
	plan        plan.Plan
	expiresDate time.Time
	payMethod   enum.PayMethod
}

func NewProfile() *Profile {
	return &Profile{
		FtcID:      uuid.New().String(),
		UnionID:    GenWxID(),
		StripeID:   GenCusID(),
		Email:      gofakeit.Email(),
		Password:   SimplePassword(),
		UserName:   gofakeit.Username(),
		Nickname:   gofakeit.Name(),
		Avatar:     GenAvatar(),
		OpenID:     GenWxID(),
		IP:         gofakeit.IPv4Address(),
		AppleSubID: GenAppleSubID(),

		kind:        reader.AccountKindFtc,
		plan:        yearlyStandard,
		expiresDate: time.Now(),
		payMethod:   enum.PayMethodAli,
	}
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

func (p Profile) Membership() Membership {
	m := Membership{
		ID:            null.StringFrom(GenerateMembershipIndex()),
		MemberID:      p.AccountID(),
		BasePlan:      p.plan.BasePlan,
		ExpireDate:    chrono.DateFrom(p.expiresDate),
		PaymentMethod: p.payMethod,
		StripeSubID:   null.String{},
		StripePlanID:  null.String{},
		AutoRenew:     false,
		Status:        SubStatusNull,
	}

	if p.payMethod == enum.PayMethodApple {
		m.AppleSubID = null.StringFrom(p.AppleSubID)
		m.AutoRenew = true
	}

	return m
}
