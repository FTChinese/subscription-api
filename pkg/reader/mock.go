// +build !production

package reader

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/google/uuid"
	"github.com/guregu/null"
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
