package wxlogin

import (
	"strings"

	"gitlab.com/ftchinese/subscription-api/util"
)

// UserInfo is a wechat user's personal information.
type UserInfo struct {
	UnionID    string `json:"unionid"`
	OpenID     string `json:"openid"`
	NickName   string `json:"nickname"`
	HeadImgURL string `json:"headimgurl"`
	// 1 for male, 2 for female, 0 for not set.
	Gender     int64    `json:"sex"`
	Country    string   `json:"country"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Privileges []string `json:"privilege"`
}

// ToWechat returns a ToWechat type from UserInfo.
func (u UserInfo) ToWechat() Wechat {
	return Wechat{
		UnionID:   u.UnionID,
		OpenID:    u.OpenID,
		NickName:  u.NickName,
		AvatarURL: u.HeadImgURL,
	}
}

// SaveUserInfo from wechat API.
func (env Env) SaveUserInfo(u UserInfo, c util.RequestClient) error {
	query := `INSERT INTO user_db.user_sns_info
	SET client_type = ?,
		unionid = ?,
		openid = ?,
		nickname = ?,
		sex = ?,
		country = ?,
		province = ?,
		city = ?,
		headimgurl = ?,
		privilege = NULLIF(?, '')
	ON DUPLICATE KEY UPDATE
		openid = ?,
		nickname = ?,
		sex = ?,
		country = ?,
		province = ?,
		city = ?,
		headimgurl = ?,
		privilege = NULLIF(?, '')`

	prvlg := strings.Join(u.Privileges, ",")
	_, err := env.DB.Exec(query,
		c.ClientType,
		u.UnionID,
		u.OpenID,
		u.NickName,
		u.Gender,
		u.Country,
		u.Province,
		u.City,
		u.HeadImgURL,
		prvlg,
		u.OpenID,
		u.NickName,
		u.Gender,
		u.Country,
		u.Province,
		u.City,
		u.HeadImgURL,
		prvlg,
	)

	if err != nil {
		logger.WithField("trace", "SaveUserInfo").Error(err)
		return err
	}

	return nil
}
