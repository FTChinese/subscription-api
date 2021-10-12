//go:build !production
// +build !production

package test

import (
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"time"

	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
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
	Password   string // Deprecated
	UserName   string
	Mobile     string
	Nickname   string
	Avatar     string
	OpenID     string
	IP         string
	AppleSubID string
	kind       enum.AccountKind
}

func NewPersona() *Persona {
	faker.SeedGoFake()

	return &Persona{
		FtcID:      uuid.New().String(),
		UnionID:    faker.GenWxID(),
		StripeID:   faker.GenCustomerID(),
		Email:      gofakeit.Email(),
		Password:   "12345678",
		UserName:   gofakeit.Username(),
		Mobile:     faker.GenPhone(),
		Nickname:   gofakeit.Name(),
		Avatar:     faker.GenAvatar(),
		AppleSubID: faker.GenAppleSubID(),
		OpenID:     faker.GenWxID(),
		IP:         gofakeit.IPv4Address(),
		kind:       enum.AccountKindFtc,
	}
}

func (p *Persona) WithAccountKind(k enum.AccountKind) *Persona {
	p.kind = k
	return p
}

func (p *Persona) WithEmailKind() *Persona {
	return p.WithAccountKind(enum.AccountKindFtc)
}

func (p *Persona) WithWxKind() *Persona {
	return p.WithAccountKind(enum.AccountKindWx)
}

func (p *Persona) WithLinkedKind() *Persona {
	return p.WithAccountKind(enum.AccountKindLinked)
}

// WithMobile set/unset mobile.
func (p *Persona) WithMobile(m string) *Persona {
	p.Mobile = m
	return p
}

func (p *Persona) AccountID() ids.UserIDs {

	var id ids.UserIDs

	switch p.kind {
	case enum.AccountKindFtc:
		id = ids.UserIDs{
			CompoundID: p.FtcID,
			FtcID:      null.StringFrom(p.FtcID),
			UnionID:    null.String{},
		}

	case enum.AccountKindWx:
		id = ids.UserIDs{
			CompoundID: p.UnionID,
			FtcID:      null.String{},
			UnionID:    null.StringFrom(p.UnionID),
		}

	case enum.AccountKindLinked:
		id = ids.UserIDs{
			CompoundID: p.FtcID,
			FtcID:      null.StringFrom(p.FtcID),
			UnionID:    null.StringFrom(p.UnionID),
		}
	}

	return id
}

func (p *Persona) EmailSignUpParams() input.EmailSignUpParams {
	return input.EmailSignUpParams{
		EmailCredentials: input.EmailCredentials{
			Email:    p.Email,
			Password: p.Password,
		},
		DeviceToken: null.String{},
		SourceURL:   "",
	}
}

func (p *Persona) MobileSignUpParams() input.MobileSignUpParams {
	return input.MobileSignUpParams{
		Mobile:      p.Mobile,
		DeviceToken: null.String{},
	}
}

func (p *Persona) EmailOnlyAccount() account.BaseAccount {
	return account.BaseAccount{
		FtcID:        p.FtcID,
		UnionID:      null.String{},
		StripeID:     null.NewString(p.StripeID, p.StripeID != ""),
		Email:        p.Email,
		Password:     "12345678",
		Mobile:       null.String{},
		UserName:     null.StringFrom(p.UserName),
		AvatarURL:    null.String{},
		IsVerified:   false,
		CampaignCode: null.String{},
	}
}

func (p *Persona) MobileOnlyAccount() account.BaseAccount {
	return account.BaseAccount{
		FtcID:        p.FtcID,
		UnionID:      null.String{},
		StripeID:     null.String{},
		Email:        account.MobileEmail(p.Mobile),
		Password:     "12345678",
		Mobile:       null.StringFrom(p.Mobile),
		UserName:     null.StringFrom(p.Mobile),
		AvatarURL:    null.String{},
		IsVerified:   false,
		CampaignCode: null.String{},
	}
}

func (p *Persona) EmailMobileAccount() account.BaseAccount {
	return account.BaseAccount{
		FtcID:        p.FtcID,
		UnionID:      null.String{},
		StripeID:     null.NewString(p.StripeID, p.StripeID != ""),
		Email:        p.Email,
		Password:     "12345678",
		Mobile:       null.StringFrom(p.Mobile),
		UserName:     null.StringFrom(p.UserName),
		AvatarURL:    null.String{},
		IsVerified:   false,
		CampaignCode: null.String{},
	}
}

func (p *Persona) MobileUpdater() account.MobileUpdater {
	return account.MobileUpdater{
		FtcID:  p.FtcID,
		Mobile: null.StringFrom(p.Mobile),
	}
}

func (p *Persona) MobileVerifier() ztsms.Verifier {
	return ztsms.NewVerifier(p.Mobile, null.StringFrom(p.FtcID))
}

func (p *Persona) MemberBuilder() MemberBuilder {
	return MemberBuilder{
		accountKind:  p.kind,
		ftcID:        p.FtcID,
		unionID:      p.UnionID,
		price:        price.Price{},
		payMethod:    0,
		expiration:   time.Time{},
		subsStatus:   0,
		autoRenewal:  false,
		addOn:        addon.AddOn{},
		iapTxID:      "",
		stripeSubsID: "",
		b2bLicID:     "",
	}
}

func (p *Persona) IAPLinkInput() apple.LinkInput {
	return apple.LinkInput{
		FtcID:        p.FtcID,
		OriginalTxID: p.AppleSubID,
	}
}

// IAPBuilder creates a linked builder instance.
func (p *Persona) IAPBuilder() IAPBuilder {
	return NewIAPBuilder(p.AppleSubID).WithFtcID(p.FtcID)
}

func (p *Persona) OrderBuilder() subs.MockOrderBuilder {
	return subs.NewMockOrderBuilder(p.FtcID)
}
