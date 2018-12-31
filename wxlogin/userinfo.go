package wxlogin

import (
	"strings"

	"github.com/guregu/null"
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

func (u UserInfo) gender() null.String {
	var g null.String
	switch u.Sex {
	case 1:
		g = null.StringFrom("M")
	case 2:
		g = null.StringFrom("F")
	}

	return g
}

// SaveUserInfo from wechat API.
// Since a user can authorize multiple times, use ON DUPLICATE to handle such situations.
func (env Env) SaveUserInfo(u UserInfo) error {
	query := `
	INSERT INTO user_db.wechat_userinfo
	SET union_id = ?,
		nickname = ?,
		avatar_url = ?,
		gender = ?,
		country = ?,
		province = ?,
		city = ?,
		privilege = NULLIF(?, '')
	ON DUPLICATE KEY UPDATE
		nickname = ?,
		avatar_url = ?,
		gender = ?,
		country = ?,
		province = ?,
		city = ?,
		privilege = NULLIF(?, '')`

	prvlg := strings.Join(u.Privileges, ",")

	_, err := env.DB.Exec(query,
		u.UnionID,
		u.NickName,
		u.AvatarURL,
		u.gender(),
		u.Country,
		u.Province,
		u.City,
		prvlg,
		u.NickName,
		u.AvatarURL,
		u.gender(),
		u.Country,
		u.Province,
		u.City,
		prvlg,
	)

	if err != nil {
		logger.WithField("trace", "SaveUserInfo").Error(err)
		return err
	}

	return nil
}

// UpdateUserInfo update data of one union id.
// This will update multiple rows if user logged in with wechat
// on different platforms.
func (env Env) UpdateUserInfo(u UserInfo) error {
	query := `
	UPDATE user_db.wechat_userinfo
	SET nickname = ?,
		gender = ?,
		country = ?,
		province = ?,
		city = ?,
		avatar_url = ?,
		privilege = NULLIF(?, '')
	WHERE union_id = ?`

	prvl := strings.Join(u.Privileges, ",")

	_, err := env.DB.Exec(query,
		u.NickName,
		u.gender(),
		u.Country,
		u.Province,
		u.City,
		u.AvatarURL,
		prvl,
		u.UnionID,
	)

	if err != nil {
		logger.WithField("trace", "UpdateUserInfo").Error(err)
		return err
	}

	return nil
}
