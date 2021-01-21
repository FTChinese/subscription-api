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
	ids         MemberID
	plan        product.Plan
	payMethod   enum.PayMethod
	expiration  time.Time
	autoRenewal bool
	reserved    ReservedDays
}

func NewMockMemberBuilder(ftcID string) MockMemberBuilder {
	if ftcID == "" {
		ftcID = uuid.New().String()
	}

	return MockMemberBuilder{
		ids: MemberID{
			CompoundID: ftcID,
			FtcID:      null.StringFrom(ftcID),
			UnionID:    null.String{},
		},
		plan:       faker.PlanStdYear.Plan,
		payMethod:  enum.PayMethodAli,
		expiration: time.Now().AddDate(0, 1, 0),
	}
}

func (b MockMemberBuilder) WithIDs(ids MemberID) MockMemberBuilder {
	b.ids = ids
	return b
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

func (b MockMemberBuilder) WithAutoRenewal(t bool) MockMemberBuilder {
	b.autoRenewal = t
	return b
}

func (b MockMemberBuilder) WithReservedDays(r ReservedDays) MockMemberBuilder {
	b.reserved = r
	return b
}

func (b MockMemberBuilder) Build() Membership {
	m := Membership{
		MemberID:      b.ids,
		Edition:       b.plan.Edition,
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    chrono.DateFrom(b.expiration),
		PaymentMethod: b.payMethod,
		FtcPlanID:     null.String{},
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   b.autoRenewal,
		Status:        0,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
		ReservedDays:  b.reserved,
	}
	switch b.payMethod {
	case enum.PayMethodAli, enum.PayMethodWx:
		m.FtcPlanID = null.StringFrom(b.plan.ID)

	case enum.PayMethodStripe:
		m.StripeSubsID = null.StringFrom(faker.GenStripeSubID())
		m.StripePlanID = null.StringFrom(faker.GenStripePlanID())
		if b.autoRenewal {
			m.Status = enum.SubsStatusActive
		} else {
			m.Status = enum.SubsStatusCanceled
		}

	case enum.PayMethodApple:
		m.AppleSubsID = null.StringFrom(faker.GenAppleSubID())
	}

	return m.Sync()
}
