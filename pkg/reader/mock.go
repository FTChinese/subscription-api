// +build !production

package reader

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/price"
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
	ids          pkg.UserIDs
	price        price.Price
	payMethod    enum.PayMethod
	expiration   time.Time
	subsStatus   enum.SubsStatus
	autoRenewal  bool
	addOn        addon.AddOn
	iapTxID      string
	stripeSubsID string
}

func NewMockMemberBuilder(ftcID string) MockMemberBuilder {
	if ftcID == "" {
		ftcID = uuid.New().String()
	}

	return MockMemberBuilder{
		ids: pkg.UserIDs{
			CompoundID: ftcID,
			FtcID:      null.StringFrom(ftcID),
			UnionID:    null.String{},
		},
		price:      price.PriceStdYear.Price,
		payMethod:  enum.PayMethodAli,
		expiration: time.Now().AddDate(0, 1, 0),
	}
}

func (b MockMemberBuilder) WithIDs(ids pkg.UserIDs) MockMemberBuilder {
	b.ids = ids
	return b
}

func (b MockMemberBuilder) WithPrice(p price.Price) MockMemberBuilder {
	b.price = p

	return b
}

func (b MockMemberBuilder) WithPayMethod(m enum.PayMethod) MockMemberBuilder {
	b.payMethod = m
	if m == enum.PayMethodStripe || m == enum.PayMethodApple {
		b.autoRenewal = true
		b.subsStatus = enum.SubsStatusActive
	}
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

func (b MockMemberBuilder) WithSubsStatus(s enum.SubsStatus) MockMemberBuilder {
	b.subsStatus = s
	return b
}

func (b MockMemberBuilder) WithAddOn(r addon.AddOn) MockMemberBuilder {
	b.addOn = r
	return b
}

func (b MockMemberBuilder) WithIapID(id string) MockMemberBuilder {
	b.iapTxID = id
	return b
}

func (b MockMemberBuilder) Build() Membership {
	m := Membership{
		UserIDs:       b.ids,
		Edition:       b.price.Edition,
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    chrono.DateFrom(b.expiration),
		PaymentMethod: b.payMethod,
		FtcPlanID:     null.String{},
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   b.autoRenewal,
		Status:        b.subsStatus,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
		AddOn:         b.addOn,
	}
	switch b.payMethod {
	case enum.PayMethodAli, enum.PayMethodWx:
		m.FtcPlanID = null.StringFrom(b.price.ID)

	case enum.PayMethodStripe:
		m.StripeSubsID = null.StringFrom(faker.GenStripeSubID())
		m.StripePlanID = null.StringFrom(faker.GenStripePlanID())

	case enum.PayMethodApple:
		if b.iapTxID == "" {
			b.iapTxID = faker.GenAppleSubID()
		}
		m.AppleSubsID = null.StringFrom(b.iapTxID)
	}

	return m.Sync()
}
