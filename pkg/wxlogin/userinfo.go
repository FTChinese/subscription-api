package wxlogin

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"strings"
)

func convertGender(s int) enum.Gender {
	switch s {
	case 1:
		return enum.GenderMale
	case 2:
		return enum.GenderFemale
	}

	return enum.GenderNull
}

func convertStrSlice(s []string) null.String {
	str := strings.Join(s, ",")
	return null.NewString(str, str != "")
}

type UserInfoShared struct {
	UnionID   string      `json:"unionid" db:"union_id"`
	OpenID    string      `json:"openid" db:"open_id"`
	NickName  null.String `json:"nickname" db:"nickname"`
	AvatarURL null.String `json:"headimgurl" db:"avatar_url"`
	Country   null.String `json:"country" db:"country"`
	Province  null.String `json:"province" db:"province"`
	City      null.String `json:"city" db:"city"`
}

// UserInfo is the response of Wechat endpoint
// /sns/userinfo?access_token=ACCESS_TOKEN&openid=OPENID
type UserInfo struct {
	UserInfoShared
	// 1 for male, 2 for female, 0 for not set.
	Sex        int      `json:"sex"`
	Privileges []string `json:"privilege"`
	RespStatus
}

func (u UserInfo) SQLSchema() UserInfoSchema {
	return UserInfoSchema{
		UserInfoShared: u.UserInfoShared,
		Gender:         convertGender(u.Sex),
		Privilege:      convertStrSlice(u.Privileges),
	}
}

type UserInfoSchema struct {
	UserInfoShared
	Gender    enum.Gender `db:"gender"`
	Privilege null.String `db:"privilege"`
}
