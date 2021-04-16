// +build !production

package wxlogin

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
)

func MockUserInfo(unionID string) UserInfoSchema {
	return UserInfoSchema{
		UserInfoShared: UserInfoShared{
			UnionID:   unionID,
			OpenID:    faker.GenWxID(),
			NickName:  null.StringFrom(gofakeit.Username()),
			AvatarURL: null.StringFrom(faker.GenAvatar()),
			Country:   null.StringFrom(gofakeit.Country()),
			Province:  null.StringFrom(gofakeit.State()),
			City:      null.StringFrom(gofakeit.City()),
		},
		Gender:    faker.RandomGender(),
		Privilege: null.String{},
	}
}
