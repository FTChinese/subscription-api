package wxlogin

import (
	"github.com/guregu/null"
	"strings"
)

// UserInfo is the response of Wechat endpoint
// /sns/userinfo?access_token=ACCESS_TOKEN&openid=OPENID
type UserInfo struct {
	UnionID   string `json:"unionid"`
	OpenID    string `json:"openid"`
	NickName  string `json:"nickname"`
	AvatarURL string `json:"headimgurl"`
	// 1 for male, 2 for female, 0 for not set.
	Sex        int      `json:"sex"`
	Country    string   `json:"country"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Privileges []string `json:"privilege"`
	RespStatus
}

// GetGender creates a nullable string from Wechat's `sex` field so  that it could saved as enum in database.
func (u UserInfo) GetGender() null.String {
	var g null.String
	switch u.Sex {
	case 1:
		g = null.StringFrom("M")
	case 2:
		g = null.StringFrom("F")
	}

	return g
}

func (u UserInfo) GetPrivilege() null.String {
	var p null.String
	str := strings.Join(u.Privileges, ",")
	if str == "" {
		return p
	}

	p = null.StringFrom(str)

	return p
}
