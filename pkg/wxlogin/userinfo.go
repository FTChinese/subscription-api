package wxlogin

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"strings"
)

// UserInfoParams is the parameter used to request userinfo.
//
type UserInfoParams struct {
	AccessToken string
	OpenID      string
}

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
	AvatarURL null.String `json:"headimgurl" db:"avatar_url"`
	City      null.String `json:"city" db:"city"`
	Country   null.String `json:"country" db:"country"`
	NickName  null.String `json:"nickname" db:"nickname"`
	Province  null.String `json:"province" db:"province"`
}

// UserInfoResponse is the response of Wechat endpoint
// /sns/userinfo?access_token=ACCESS_TOKEN&openid=OPENID
type UserInfoResponse struct {
	UserInfoShared
	// 1 for male, 2 for female, 0 for not set.
	Sex        int      `json:"sex"`
	Privileges []string `json:"privilege"`
	RespStatus
}

// UserInfoSchema standardizes wechat userinfo response
type UserInfoSchema struct {
	UserInfoShared
	Gender     enum.Gender `json:"gender" db:"gender"`
	Privilege  null.String `json:"privilege" db:"privilege"`
	CreatedUTC chrono.Time `json:"createdUtc" db:"created_utc"`
	UpdateUTC  chrono.Time `json:"updateUtc" db:"updated_utc"`
}

func NewUserInfo(u UserInfoResponse) UserInfoSchema {
	return UserInfoSchema{
		UserInfoShared: u.UserInfoShared,
		Gender:         convertGender(u.Sex),
		Privilege:      convertStrSlice(u.Privileges),
		CreatedUTC:     chrono.TimeNow(), // Ignored upon updating.
		UpdateUTC:      chrono.TimeNow(),
	}
}
