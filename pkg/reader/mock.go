// +build !production

package reader

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"time"
)

type MockMemberBuilder struct {
	accountKind  enum.AccountKind
	ftcID        string
	unionID      string
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
		accountKind: enum.AccountKindFtc,
		ftcID:       ftcID,
		unionID:     faker.GenWxID(),
		price:       price.MockPriceStdYear.Price,
		payMethod:   enum.PayMethodAli,
		expiration:  time.Now().AddDate(0, 1, 0),
	}
}

func NewMockMemberBuilderV2(k enum.AccountKind) MockMemberBuilder {
	return MockMemberBuilder{
		accountKind:  k,
		ftcID:        uuid.New().String(),
		unionID:      faker.GenWxID(),
		price:        price.MockPriceStdYear.Price,
		payMethod:    enum.PayMethodAli,
		expiration:   time.Now().AddDate(0, 1, 0),
		subsStatus:   0,
		autoRenewal:  false,
		addOn:        addon.AddOn{},
		iapTxID:      "",
		stripeSubsID: "",
	}
}

func (b MockMemberBuilder) WithAccountKind(k enum.AccountKind) MockMemberBuilder {
	b.accountKind = k
	return b
}

func (b MockMemberBuilder) WithIDs(ids pkg.UserIDs) MockMemberBuilder {
	b.ftcID = ids.FtcID.String
	b.unionID = ids.UnionID.String
	return b
}

func (b MockMemberBuilder) WithFtcID(id string) MockMemberBuilder {
	b.ftcID = id
	return b
}

func (b MockMemberBuilder) WithWxID(id string) MockMemberBuilder {
	b.unionID = id
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
	var ids pkg.UserIDs
	switch b.accountKind {
	case enum.AccountKindFtc:
		ids = pkg.UserIDs{
			CompoundID: b.ftcID,
			FtcID:      null.StringFrom(b.ftcID),
			UnionID:    null.String{},
		}
	case enum.AccountKindWx:
		ids = pkg.UserIDs{
			CompoundID: b.unionID,
			FtcID:      null.String{},
			UnionID:    null.StringFrom(b.unionID),
		}
	case enum.AccountKindLinked:
		ids = pkg.UserIDs{
			CompoundID: b.ftcID,
			FtcID:      null.StringFrom(b.ftcID),
			UnionID:    null.StringFrom(b.unionID),
		}
	}

	m := Membership{
		UserIDs:       ids,
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
