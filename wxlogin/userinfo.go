package wxlogin

import (
	"strings"

	"gitlab.com/ftchinese/subscription-api/util"
)

// UserInfo is a wechat user's personal information.
type UserInfo struct {
	OpenID     string   `json:"openid"`
	NickName   string   `json:"nickname"`
	Gender     string   `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privileges []string `json:"privilege"`
	UnionID    string   `json:"unionid"`
}

// SaveUserInfo from wechat API.
func (env Env) SaveUserInfo(u UserInfo, c util.RequestClient) error {
	query := `INSERT INTO user_db.user_sns_info
	SET client_type = ?
		unionid = ?,
		openid = ?,
		nickname = ?,
		sex = ?,
		country = ?,
		province = ?,
		city = ?,
		headimgurl = ?,
		privilege = ?
	ON DUPLICATE KEY UPDATE
		openid = ?,
		nickname = ?,
		sex = ?,
		country = ?,
		province = ?,
		city = ?,
		headimgurl = ?,
		privilege = ?`

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
