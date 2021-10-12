package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"time"
)

type MemberBuilder struct {
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
	b2bLicID     string
}

func (b MemberBuilder) WithAccountKind(k enum.AccountKind) MemberBuilder {
	b.accountKind = k
	return b
}

func (b MemberBuilder) WithIDs(ids ids.UserIDs) MemberBuilder {
	b.ftcID = ids.FtcID.String
	b.unionID = ids.UnionID.String
	return b
}

func (b MemberBuilder) WithFtcID(id string) MemberBuilder {
	b.ftcID = id
	return b
}

func (b MemberBuilder) WithWxID(id string) MemberBuilder {
	b.unionID = id
	return b
}

// WithPayMethod sets payment method.
// Deprecated.
func (b MemberBuilder) WithPayMethod(m enum.PayMethod) MemberBuilder {
	b.payMethod = m
	if m == enum.PayMethodStripe || m == enum.PayMethodApple {
		b.autoRenewal = true
		b.subsStatus = enum.SubsStatusActive
	}
	return b
}

func (b MemberBuilder) WithAlipay() MemberBuilder {
	b.payMethod = enum.PayMethodAli
	return b
}

func (b MemberBuilder) WithWx() MemberBuilder {
	b.payMethod = enum.PayMethodWx
	return b
}

// WithPrice sets the subscription plan for payment method alipay or wechat.
// Call this together with WithAlipay or WithWx
func (b MemberBuilder) WithPrice(p price.Price) MemberBuilder {
	b.price = p

	return b
}

func (b MemberBuilder) WithStripe(subsID string) MemberBuilder {
	if subsID == "" {
		subsID = faker.GenStripeSubID()
	}

	b.payMethod = enum.PayMethodStripe
	b.stripeSubsID = subsID
	b.autoRenewal = true
	b.subsStatus = enum.SubsStatusActive

	return b
}

func (b MemberBuilder) WithApple(txID string) MemberBuilder {
	if txID == "" {
		txID = faker.GenAppleSubID()
	}

	b.iapTxID = txID
	b.autoRenewal = true

	return b
}

func (b MemberBuilder) WithB2B(licID string) MemberBuilder {
	if licID == "" {
		licID = faker.GenLicenceID()
	}

	b.b2bLicID = licID
	b.payMethod = enum.PayMethodB2B

	return b
}

func (b MemberBuilder) WithExpiration(t time.Time) MemberBuilder {
	b.expiration = t
	return b
}

// WithAutoRenewal switches auto renewal status.
// Deprecated.
func (b MemberBuilder) WithAutoRenewal(t bool) MemberBuilder {
	b.autoRenewal = t
	return b
}

func (b MemberBuilder) WithAutoRenewOn() MemberBuilder {
	b.autoRenewal = true
	return b
}

func (b MemberBuilder) WithAutoRenewOff() MemberBuilder {
	b.autoRenewal = false
	return b
}

func (b MemberBuilder) WithSubsStatus(s enum.SubsStatus) MemberBuilder {
	b.subsStatus = s
	return b
}

func (b MemberBuilder) WithAddOn(r addon.AddOn) MemberBuilder {
	b.addOn = r
	return b
}

// WithIapID sets iap subscription id.
// Deprecated.
func (b MemberBuilder) WithIapID(id string) MemberBuilder {
	b.iapTxID = id
	return b
}

func (b MemberBuilder) Build() reader.Membership {
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

	m := reader.Membership{
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
		m.StripePlanID = null.StringFrom(faker.GenStripePlanID())

	case enum.PayMethodApple:
		m.AppleSubsID = null.StringFrom(b.iapTxID)

	case enum.PayMethodB2B:
		m.B2BLicenceID = null.StringFrom(b.b2bLicID)
	}

	return m.Sync()
}
