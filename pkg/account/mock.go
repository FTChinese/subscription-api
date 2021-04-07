// +build !production

package account

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/google/uuid"
	"github.com/guregu/null"
)

type MockFtcAccountBuilder struct {
	kind     enum.AccountKind
	ftcID    string
	wxID     string
	stripeID string
	email    string
	mobile   string
}

func NewMockFtcAccountBuilder(kind enum.AccountKind) MockFtcAccountBuilder {

	faker.SeedGoFake()

	return MockFtcAccountBuilder{
		kind:     kind,
		ftcID:    uuid.New().String(),
		wxID:     faker.GenWxID(),
		stripeID: faker.GenCustomerID(),
		email:    gofakeit.Email(),
		mobile:   gofakeit.Phone(),
	}
}

func (b MockFtcAccountBuilder) WithFtcID(id string) MockFtcAccountBuilder {
	b.ftcID = id
	return b
}

func (b MockFtcAccountBuilder) WithWxID(id string) MockFtcAccountBuilder {
	b.wxID = id
	return b
}

func (b MockFtcAccountBuilder) WithStripeID(id string) MockFtcAccountBuilder {
	b.stripeID = id
	return b
}

func (b MockFtcAccountBuilder) WithEmail(e string) MockFtcAccountBuilder {
	b.email = e
	return b
}

func (b MockFtcAccountBuilder) WithMobile(m string) MockFtcAccountBuilder {
	b.mobile = m
	return b
}

func (b MockFtcAccountBuilder) Build() BaseAccount {
	switch b.kind {
	case enum.AccountKindFtc:
		return BaseAccount{
			FtcID:      b.ftcID,
			UnionID:    null.String{},
			StripeID:   null.NewString(b.stripeID, b.stripeID != ""),
			Email:      b.email,
			Password:   "12345678",
			Mobile:     null.NewString(b.mobile, b.mobile != ""),
			UserName:   null.StringFrom(gofakeit.Username()),
			AvatarURL:  null.String{},
			IsVerified: false,
		}

	case enum.AccountKindWx:
		return BaseAccount{
			FtcID:      "",
			UnionID:    null.StringFrom(b.wxID),
			StripeID:   null.NewString(b.stripeID, b.stripeID != ""),
			Email:      b.email,
			Password:   "12345678",
			Mobile:     null.NewString(b.mobile, b.mobile != ""),
			UserName:   null.StringFrom(gofakeit.Username()),
			AvatarURL:  null.String{},
			IsVerified: false,
		}

	case enum.AccountKindLinked:
		return BaseAccount{
			FtcID:      b.ftcID,
			UnionID:    null.StringFrom(b.wxID),
			StripeID:   null.NewString(b.stripeID, b.stripeID != ""),
			Email:      b.email,
			Password:   "12345678",
			Mobile:     null.NewString(b.mobile, b.mobile != ""),
			UserName:   null.StringFrom(gofakeit.Username()),
			AvatarURL:  null.String{},
			IsVerified: false,
		}

	default:
		return BaseAccount{}
	}
}
