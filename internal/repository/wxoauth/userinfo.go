package wxoauth

import (
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
)

// SaveWxUser from wechat API.
// Since a user can authorize multiple times, use ON DUPLICATE to handle unique key constraint.
func (env Env) SaveWxUser(u wxlogin.UserInfo) error {

	_, err := env.db.Exec(wxlogin.StmtInsertUserInfo,
		u.UnionID,
		u.NickName,
		u.AvatarURL,
		u.GetGender(),
		u.Country,
		u.Province,
		u.City,
		u.GetPrivilege(),
		u.NickName,
		u.AvatarURL,
		u.GetGender(),
		u.Country,
		u.Province,
		u.City,
		u.GetPrivilege(),
	)

	if err != nil {
		return err
	}

	return nil
}

// UpdateWxUser update data of one union id.
func (env Env) UpdateWxUser(u wxlogin.UserInfo) error {

	_, err := env.db.Exec(wxlogin.StmtUpdateUserInfo,
		u.NickName,
		u.GetGender(),
		u.Country,
		u.Province,
		u.City,
		u.AvatarURL,
		u.GetPrivilege(),
		u.UnionID,
	)

	if err != nil {
		return err
	}

	return nil
}
