//go:build !production
// +build !production

package reader

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"time"
)

type MockMemberBuilder struct {
	accountKind  enum.AccountKind
	ftcID        string
	unionID      string
	price        price.FtcPrice
	payMethod    enum.PayMethod
	expiration   time.Time
	subsStatus   enum.SubsStatus
	autoRenewal  bool
	addOn        addon.AddOn
	iapTxID      string
	stripeSubsID string
	b2bLicID     string
}

// NewMockMemberBuilder creates a new membership builder.
// Deprecated
func NewMockMemberBuilder(ftcID string) MockMemberBuilder {
	if ftcID == "" {
		ftcID = uuid.New().String()
	}

	return MockMemberBuilder{
		accountKind: enum.AccountKindFtc,
		ftcID:       ftcID,
		unionID:     faker.GenWxID(),
		price:       pw.MockPwPriceStdYear.FtcPrice,
		payMethod:   enum.PayMethodAli,
		expiration:  time.Now().AddDate(0, 1, 0),
	}
}

func NewMockMemberBuilderV2(k enum.AccountKind) MockMemberBuilder {
	return MockMemberBuilder{
		accountKind:  k,
		ftcID:        uuid.New().String(),
		unionID:      faker.GenWxID(),
		price:        pw.MockPwPriceStdYear.FtcPrice,
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

func (b MockMemberBuilder) WithIDs(ids ids.UserIDs) MockMemberBuilder {
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

// WithPayMethod sets payment method.
// Deprecated.
func (b MockMemberBuilder) WithPayMethod(m enum.PayMethod) MockMemberBuilder {
	b.payMethod = m
	if m == enum.PayMethodStripe || m == enum.PayMethodApple {
		b.autoRenewal = true
		b.subsStatus = enum.SubsStatusActive
	}
	return b
}

func (b MockMemberBuilder) WithAlipay() MockMemberBuilder {
	b.payMethod = enum.PayMethodAli
	return b
}

func (b MockMemberBuilder) WithWx() MockMemberBuilder {
	b.payMethod = enum.PayMethodWx
	return b
}

// WithPrice sets the subscription plan for payment method alipay or wechat.
// Call this together with WithAlipay or WithWx
func (b MockMemberBuilder) WithPrice(p price.FtcPrice) MockMemberBuilder {
	b.price = p

	return b
}

func (b MockMemberBuilder) WithStripe(subsID string) MockMemberBuilder {
	if subsID == "" {
		subsID = faker.GenStripeSubID()
	}

	b.payMethod = enum.PayMethodStripe
	b.stripeSubsID = subsID
	b.autoRenewal = true
	b.subsStatus = enum.SubsStatusActive

	return b
}

func (b MockMemberBuilder) WithApple(txID string) MockMemberBuilder {
	if txID == "" {
		txID = faker.GenAppleSubID()
	}

	b.iapTxID = txID
	b.autoRenewal = true

	return b
}

func (b MockMemberBuilder) WithB2B(licID string) MockMemberBuilder {
	if licID == "" {
		licID = faker.GenLicenceID()
	}

	b.b2bLicID = licID
	b.payMethod = enum.PayMethodB2B

	return b
}

func (b MockMemberBuilder) WithExpiration(t time.Time) MockMemberBuilder {
	b.expiration = t
	return b
}

// WithAutoRenewal switches auto renewal status.
// Deprecated.
func (b MockMemberBuilder) WithAutoRenewal(t bool) MockMemberBuilder {
	b.autoRenewal = t
	return b
}

func (b MockMemberBuilder) WithAutoRenewOn() MockMemberBuilder {
	b.autoRenewal = true
	return b
}

func (b MockMemberBuilder) WithAutoRenewOff() MockMemberBuilder {
	b.autoRenewal = false
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

// WithIapID sets iap subscription id.
// Deprecated.
func (b MockMemberBuilder) WithIapID(id string) MockMemberBuilder {
	b.iapTxID = id
	return b
}

func (b MockMemberBuilder) Build() Membership {
	var userIDs ids.UserIDs
	switch b.accountKind {
	case enum.AccountKindFtc:
		userIDs = ids.UserIDs{
			CompoundID: b.ftcID,
			FtcID:      null.StringFrom(b.ftcID),
			UnionID:    null.String{},
		}
	case enum.AccountKindWx:
		userIDs = ids.UserIDs{
			CompoundID: b.unionID,
			FtcID:      null.String{},
			UnionID:    null.StringFrom(b.unionID),
		}
	case enum.AccountKindLinked:
		userIDs = ids.UserIDs{
			CompoundID: b.ftcID,
			FtcID:      null.StringFrom(b.ftcID),
			UnionID:    null.StringFrom(b.unionID),
		}
	}

	m := Membership{
		UserIDs:       userIDs,
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
		m.StripeSubsID = null.StringFrom(b.stripeSubsID)
		m.StripePlanID = null.StringFrom(faker.GenStripePriceID())

	case enum.PayMethodApple:
		m.AppleSubsID = null.StringFrom(b.iapTxID)

	case enum.PayMethodB2B:
		m.B2BLicenceID = null.StringFrom(b.b2bLicID)
	}

	return m.Sync()
}
