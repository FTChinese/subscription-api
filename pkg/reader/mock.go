// +build !production

package reader

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"time"
)

func MockNewFtcAccount(kind enum.AccountKind) FtcAccount {
	switch kind {
	case enum.AccountKindFtc:
		return FtcAccount{
			FtcID:    uuid.New().String(),
			UnionID:  null.String{},
			StripeID: null.StringFrom(faker.GenCustomerID()),
			Email:    gofakeit.Email(),
			UserName: null.StringFrom(gofakeit.Username()),
		}

	case enum.AccountKindWx:
		return FtcAccount{
			FtcID:    "",
			UnionID:  null.StringFrom(faker.GenWxID()),
			StripeID: null.String{},
			Email:    "",
			UserName: null.String{},
		}

	case enum.AccountKindLinked:
		return FtcAccount{
			FtcID:    uuid.New().String(),
			UnionID:  null.StringFrom(faker.GenWxID()),
			StripeID: null.StringFrom(faker.GenCustomerID()),
			Email:    gofakeit.Email(),
			UserName: null.StringFrom(gofakeit.Username()),
		}
	}

	return FtcAccount{}
}

type MockMemberBuilder struct {
	ftcID      string
	plan       product.Plan
	payMethod  enum.PayMethod
	expiration time.Time
}

func NewMockMemberBuilder(ftcID string) MockMemberBuilder {
	if ftcID == "" {
		ftcID = uuid.New().String()
	}

	return MockMemberBuilder{
		ftcID:      ftcID,
		plan:       faker.PlanStdYear.Plan,
		payMethod:  enum.PayMethodAli,
		expiration: time.Now().AddDate(0, 1, 0),
	}
}

func (b MockMemberBuilder) WithPlan(p product.Plan) MockMemberBuilder {
	b.plan = p

	return b
}

func (b MockMemberBuilder) WithPayMethod(m enum.PayMethod) MockMemberBuilder {
	b.payMethod = m
	return b
}
func (b MockMemberBuilder) WithExpiration(t time.Time) MockMemberBuilder {
	b.expiration = t
	return b
}

func (b MockMemberBuilder) Build() Membership {
	m := Membership{
		MemberID: MemberID{
			CompoundID: b.ftcID,
			FtcID:      null.StringFrom(b.ftcID),
		},
		Edition:       b.plan.Edition,
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    chrono.DateFrom(b.expiration),
		PaymentMethod: b.payMethod,
		FtcPlanID:     null.String{},
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        0,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
		ReservedDays:  ReservedDays{},
	}
	switch b.payMethod {
	case enum.PayMethodAli, enum.PayMethodWx:
		m.FtcPlanID = null.StringFrom(b.plan.ID)

	case enum.PayMethodStripe:
		m.StripeSubsID = null.StringFrom(faker.GenStripeSubID())
		m.StripePlanID = null.StringFrom(faker.GenStripePlanID())
		m.AutoRenewal = true
		m.Status = enum.SubsStatusActive

	case enum.PayMethodApple:
		m.AppleSubsID = null.StringFrom(faker.GenAppleSubID())
		m.AutoRenewal = true
	}

	return m.Sync()
}
